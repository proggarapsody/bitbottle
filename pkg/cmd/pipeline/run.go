package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPipelineRun(f *factory.Factory) *cobra.Command {
	var branch string
	var hostname string

	cmd := &cobra.Command{
		Use:   "run PROJECT/REPO",
		Short: "Trigger a pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
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

			pl, err := pc.RunPipeline(ref.Project, ref.Slug, backend.RunPipelineInput{Branch: branch})
			if err != nil {
				return err
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "Pipeline #%d triggered on %s (state: %s)\n", pl.BuildNumber, pl.RefName, pl.State)
			if pl.WebURL != "" {
				fmt.Fprintf(out, "URL: %s\n", pl.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&branch, "branch", "", "Branch to run pipeline on (required)")
	_ = cmd.MarkFlagRequired("branch")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
