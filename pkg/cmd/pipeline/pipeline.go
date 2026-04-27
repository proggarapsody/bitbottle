package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPipeline(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Manage Bitbucket Pipelines (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdPipelineList(f))
	cmd.AddCommand(NewCmdPipelineView(f))
	cmd.AddCommand(NewCmdPipelineRun(f))
	return cmd
}
