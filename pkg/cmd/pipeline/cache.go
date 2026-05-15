package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdCache "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/cache"
)

// NewCmdPipelineCache builds the `pipeline cache` subgroup command.
func NewCmdPipelineCache(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage Cloud pipeline caches",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(cmdCache.NewCmdList(f, nil))
	cmd.AddCommand(cmdCache.NewCmdDelete(f, nil))
	return cmd
}
