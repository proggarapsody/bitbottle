package view

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/shared"
)

// Options holds parsed flags for `pipeline view`.
type Options struct {
	Hostname   string
	JSONFields string
	JQExpr     string
	Web        bool

	// Args[0] = PROJECT/REPO, Args[1] = UUID
	Args []string
}

// NewCmdView builds the `pipeline view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO UUID",
		Short: "View a pipeline",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Web, "web", false, "Open in browser")
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineClient(client, ref.Host)
	if err != nil {
		return err
	}
	pl, err := pc.GetPipeline(ref.Project, ref.Slug, opts.Args[1])
	if err != nil {
		return err
	}

	if opts.Web {
		if pl.WebURL == "" {
			return fmt.Errorf("no web URL available for this pipeline")
		}
		return f.Browser.Browse(pl.WebURL)
	}

	if opts.JSONFields != "" || opts.JQExpr != "" {
		printer := shared.PipelineFields(f, opts.JSONFields, opts.JQExpr)
		printer.SetSingleItem()
		printer.AddItem(pl)
		return printer.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "#%d %s\n", pl.BuildNumber, pl.State)
	fmt.Fprintf(out, "Branch:   %s\n", pl.RefName)
	fmt.Fprintf(out, "Duration: %ds\n", pl.Duration)
	if pl.WebURL != "" {
		fmt.Fprintf(out, "URL:      %s\n", pl.WebURL)
	}
	return nil
}
