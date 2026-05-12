package report

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// SetOptions holds parsed flags for `code-insights report set`.
type SetOptions struct {
	Hostname   string
	Title      string
	Result     string
	ReportType string
	Details    string
	Reporter   string
	Link       string
	LogoURL    string
	Data       []string // each "K=V:TYPE"
	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdSet builds the `code-insights report set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*SetOptions) error) *cobra.Command {
	opts := &SetOptions{}
	cmd := &cobra.Command{
		Use:   "set PROJECT/REPO HASH KEY",
		Short: "Create or update (upsert) a Code Insights report",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Title, "title", "", "Report title (required)")
	cmd.Flags().StringVar(&opts.Result, "result", "", "Report result: PASS, FAIL, or PENDING (required)")
	cmd.Flags().StringVar(&opts.ReportType, "report-type", "", "Report type: TESTING, COVERAGE, BUG, SECURITY, DUPLICATION, DEPENDENCY")
	cmd.Flags().StringVar(&opts.Details, "details", "", "Human-readable details about the report")
	cmd.Flags().StringVar(&opts.Reporter, "reporter", "", "Name of the tool/reporter")
	cmd.Flags().StringVar(&opts.Link, "link", "", "URL linking to the full report")
	cmd.Flags().StringVar(&opts.LogoURL, "logo-url", "", "URL of the tool's logo")
	cmd.Flags().StringArrayVar(&opts.Data, "data", nil, "Data point in the form TITLE=VALUE:TYPE (repeatable)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("result")
	return cmd
}

// parseDatum parses "TITLE=VALUE:TYPE" into a CodeInsightsReportDatum.
func parseDatum(s string) (backend.CodeInsightsReportDatum, error) {
	// Find last ":" which separates type
	colonIdx := strings.LastIndex(s, ":")
	if colonIdx < 0 {
		return backend.CodeInsightsReportDatum{}, fmt.Errorf("invalid --data %q: expected TITLE=VALUE:TYPE", s)
	}
	typ := s[colonIdx+1:]
	kv := s[:colonIdx]
	eqIdx := strings.Index(kv, "=")
	if eqIdx < 0 {
		return backend.CodeInsightsReportDatum{}, fmt.Errorf("invalid --data %q: expected TITLE=VALUE:TYPE", s)
	}
	title := kv[:eqIdx]
	value := kv[eqIdx+1:]
	if title == "" || typ == "" {
		return backend.CodeInsightsReportDatum{}, fmt.Errorf("invalid --data %q: title and type must not be empty", s)
	}
	return backend.CodeInsightsReportDatum{Title: title, Type: strings.ToUpper(typ), Value: value}, nil
}

func setRun(f *factory.Factory, opts *SetOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	hash := opts.Args[1]
	key := opts.Args[2]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}
	data := make([]backend.CodeInsightsReportDatum, 0, len(opts.Data))
	for _, d := range opts.Data {
		datum, err := parseDatum(d)
		if err != nil {
			return err
		}
		data = append(data, datum)
	}
	in := backend.CodeInsightsReportInput{
		Title:      opts.Title,
		Result:     strings.ToUpper(opts.Result),
		ReportType: strings.ToUpper(opts.ReportType),
		Details:    opts.Details,
		Reporter:   opts.Reporter,
		Link:       opts.Link,
		LogoURL:    opts.LogoURL,
		Data:       data,
	}
	r, err := ci.SetReport(ref.Project, ref.Slug, hash, key, in)
	if err != nil {
		return err
	}
	p := reportFields(f, format.OutputConfig{})
	p.SetSingleItem()
	p.AddItem(r)
	return p.Render()
}
