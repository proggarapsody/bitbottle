package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdSchedule "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/schedule"
)

// NewCmdPipelineSchedule builds the `pipeline schedule` subgroup command.
func NewCmdPipelineSchedule(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage Cloud pipeline schedules",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(cmdSchedule.NewCmdList(f, nil))
	cmd.AddCommand(cmdSchedule.NewCmdCreate(f, nil))
	cmd.AddCommand(cmdSchedule.NewCmdDelete(f, nil))
	return cmd
}
