// Package user is the `admin user` subcommand tree.
package user

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdUser builds the `admin user` command tree.
func NewCmdUser(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Bitbucket Server users",
	}
	cmd.AddCommand(NewCmdUserList(f, nil))
	cmd.AddCommand(NewCmdUserActivate(f, nil))
	cmd.AddCommand(NewCmdUserDeactivate(f, nil))
	cmd.AddCommand(NewCmdUserRename(f, nil))
	return cmd
}
