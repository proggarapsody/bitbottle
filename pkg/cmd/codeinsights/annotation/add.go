package annotation

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// AddOptions holds parsed flags for `code-insights annotation add`.
type AddOptions struct {
	Hostname string
	FromJSON string // path or "-" for stdin; mutually exclusive with single flags

	// Single-annotation flags
	Path       string
	Line       int
	Severity   string
	Type       string
	Message    string
	ExternalID string
	Link       string

	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdAdd builds the `code-insights annotation add` cobra command.
func NewCmdAdd(f *factory.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO] HASH KEY",
		Short: "Add Code Insights annotations to a report (single or bulk)",
		Long: `Add one or more Code Insights annotations to a report.

Single annotation: supply --path, --line, --severity, --type, --message.

Bulk upload from JSON (array of annotation objects):
  bitbottle code-insights annotation add PROJ/REPO HASH KEY --from-json @annotations.json
  bitbottle code-insights annotation add PROJ/REPO HASH KEY --from-json -   # stdin`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return addRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.FromJSON, "from-json", "", "Read annotations from JSON file path or \"-\" for stdin (bulk mode)")
	cmd.Flags().StringVar(&opts.Path, "path", "", "File path (single mode)")
	cmd.Flags().IntVar(&opts.Line, "line", 0, "Line number (single mode)")
	cmd.Flags().StringVar(&opts.Severity, "severity", "", "Severity: LOW, MEDIUM, HIGH, CRITICAL (single mode)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Type: VULNERABILITY, CODE_SMELL, BUG (single mode)")
	cmd.Flags().StringVar(&opts.Message, "message", "", "Annotation message (single mode)")
	cmd.Flags().StringVar(&opts.ExternalID, "external-id", "", "External identifier (single mode, optional)")
	cmd.Flags().StringVar(&opts.Link, "link", "", "URL for more details (optional)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func addRun(f *factory.Factory, opts *AddOptions) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 2)
	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	hash := rest[0]
	key := rest[1]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}

	var anns []backend.CodeInsightsAnnotationInput
	if opts.FromJSON != "" {
		anns, err = readAnnotationsFromJSON(f, opts.FromJSON)
		if err != nil {
			return fmt.Errorf("--from-json: %w", err)
		}
	} else {
		if opts.Path == "" || opts.Message == "" {
			return fmt.Errorf("either --from-json or both --path and --message are required")
		}
		anns = []backend.CodeInsightsAnnotationInput{
			{
				Path:       opts.Path,
				Line:       opts.Line,
				Severity:   strings.ToUpper(opts.Severity),
				Type:       strings.ToUpper(opts.Type),
				Message:    opts.Message,
				ExternalID: opts.ExternalID,
				Link:       opts.Link,
			},
		}
	}

	if err := ci.AddAnnotations(ref.Project, ref.Slug, hash, key, anns); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Added %d annotation(s) to report %q\n", len(anns), key)
	return nil
}

func readAnnotationsFromJSON(f *factory.Factory, src string) ([]backend.CodeInsightsAnnotationInput, error) {
	var r io.Reader
	if src == "-" {
		r = f.IOStreams.In
	} else {
		path := strings.TrimPrefix(src, "@")
		fh, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer func() { _ = fh.Close() }()
		r = fh
	}
	var anns []backend.CodeInsightsAnnotationInput
	if err := json.NewDecoder(r).Decode(&anns); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return anns, nil
}
