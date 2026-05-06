package run

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline run`.
type Options struct {
	Hostname string
	Branch   string

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdRun builds the `pipeline run` cobra command.
func NewCmdRun(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "run PROJECT/REPO",
		Short: "Trigger a pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch to run pipeline on (required)")
	_ = cmd.MarkFlagRequired("branch")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runRun(f *factory.Factory, opts *Options) error {
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
	pl, err := pc.RunPipeline(ref.Project, ref.Slug, backend.RunPipelineInput{Branch: opts.Branch})
	if err != nil {
		return err
	}
	out := f.IOStreams.Out
	fmt.Fprintf(out, "Pipeline #%d triggered on %s (state: %s)\n", pl.BuildNumber, pl.RefName, pl.State)
	if pl.WebURL != "" {
		fmt.Fprintf(out, "URL: %s\n", pl.WebURL)
	}
	return nil
}
