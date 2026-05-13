package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/list"
	cmdLogs "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/logs"
	cmdRun "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/run"
	cmdSteps "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/steps"
	cmdTrigger "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/trigger"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/view"
	cmdWatch "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/watch"
)

// NewCmdPipeline builds the root `pipeline` command. Subcommands live in their
// own subpackages (gh-CLI style) so each command's surface is testable in
// isolation.
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
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdView.NewCmdView(f, nil))
	cmd.AddCommand(cmdRun.NewCmdRun(f, nil))
	cmd.AddCommand(cmdSteps.NewCmdSteps(f, nil))
	cmd.AddCommand(cmdTrigger.NewCmdTrigger(f, nil))
	cmd.AddCommand(cmdLogs.NewCmdLogs(f, nil))
	cmd.AddCommand(cmdWatch.NewCmdWatch(f, nil))
	return cmd
}
