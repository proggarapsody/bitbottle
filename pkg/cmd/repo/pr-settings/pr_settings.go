// Package prsettings implements the `repo pr-settings` command group.
package prsettings

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRSettings builds the `repo pr-settings` group command.
func NewCmdPRSettings(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr-settings",
		Short: "Manage pull request gate settings for a repository",
	}
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdSet(f))
	return cmd
}
