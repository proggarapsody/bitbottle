// Package settings implements the `repo hook settings` command group.
package settings

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSettings builds the `repo hook settings` group command.
func NewCmdSettings(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage hook script settings",
	}
	cmd.AddCommand(NewCmdSettingsGet(f))
	cmd.AddCommand(NewCmdSettingsSet(f))
	return cmd
}
