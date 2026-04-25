package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPipeline(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage Bitbucket Pipelines (Cloud only)",
	}
	cmd.AddCommand(NewCmdPipelineList(f))
	cmd.AddCommand(NewCmdPipelineView(f))
	cmd.AddCommand(NewCmdPipelineRun(f))
	return cmd
}
