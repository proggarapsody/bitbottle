package trigger

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline trigger`.
type Options struct {
	Hostname  string
	Branch    string
	Variables []string // raw "key=value" strings from --variable

	// Args[0] = PROJECT/REPO (optional)
	Args []string
}

// NewCmdTrigger builds the `pipeline trigger` cobra command.
func NewCmdTrigger(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "trigger [PROJECT/REPO]",
		Short: "Trigger a Bitbucket Cloud pipeline",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runTrigger(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch to trigger the pipeline on (defaults to current git branch)")
	cmd.Flags().StringArrayVar(&opts.Variables, "variable", nil, "Pipeline variable as key=value (repeatable)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runTrigger(f *factory.Factory, cmd *cobra.Command, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}

	// Default branch to current git branch when not explicitly set.
	branch := opts.Branch
	if branch == "" {
		g := git.New(f.GitRunner())
		branch, err = g.CurrentBranch()
		if err != nil || branch == "" {
			return fmt.Errorf("--branch not set and could not detect current git branch: %w", err)
		}
	}

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	tc, err := backend.AsPipelineTriggerClient(client, ref.Host)
	if err != nil {
		return err
	}

	vars, err := parseVariables(opts.Variables)
	if err != nil {
		return err
	}

	result, err := tc.TriggerPipeline(ref.Project, ref.Slug, backend.PipelineTriggerInput{
		Branch:    branch,
		Variables: vars,
	})
	if err != nil {
		return err
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.PipelineTriggerResult](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.SetSingleItem()
		p.AddField(format.Field[backend.PipelineTriggerResult]{Name: "uuid", Header: "UUID", Extract: func(r backend.PipelineTriggerResult) any { return r.UUID }})
		p.AddField(format.Field[backend.PipelineTriggerResult]{Name: "state", Header: "STATE", Extract: func(r backend.PipelineTriggerResult) any { return r.State }})
		p.AddField(format.Field[backend.PipelineTriggerResult]{Name: "link", Header: "LINK", Extract: func(r backend.PipelineTriggerResult) any { return r.Link }})
		p.AddItem(result)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Pipeline triggered (UUID: %s, state: %s)\n", result.UUID, result.State)
	if result.Link != "" {
		fmt.Fprintf(out, "Link: %s\n", result.Link)
	}
	return nil
}

// parseVariables converts "key=value" strings into PipelineVariable slice.
func parseVariables(raw []string) ([]backend.PipelineVariable, error) {
	vars := make([]backend.PipelineVariable, 0, len(raw))
	for _, kv := range raw {
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid --variable %q: expected key=value format", kv)
		}
		vars = append(vars, backend.PipelineVariable{
			Key:   kv[:idx],
			Value: kv[idx+1:],
		})
	}
	return vars, nil
}
