// Package extension implements the `bitbottle extension` command group.
package extension

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdExtension)
}

// NewCmdExtension returns the root `extension` command with all subcommands.
func NewCmdExtension(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Short:   "Manage bitbottle extensions",
		Aliases: []string{"ext"},
	}
	cmd.AddCommand(NewCmdInstall(f))
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdUpgrade(f))
	cmd.AddCommand(NewCmdRemove(f))
	cmd.AddCommand(NewCmdExec(f))
	return cmd
}
