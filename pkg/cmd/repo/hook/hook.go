// Package hook implements the `repo hook` command group.
package hook

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/hook/settings"
)

// NewCmdHook builds the `repo hook` group command.
func NewCmdHook(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage repo plugin hook scripts",
	}
	cmd.AddCommand(NewCmdHookList(f))
	cmd.AddCommand(NewCmdHookView(f))
	cmd.AddCommand(NewCmdHookEnable(f))
	cmd.AddCommand(NewCmdHookDisable(f))
	cmd.AddCommand(settings.NewCmdSettings(f))
	return cmd
}
