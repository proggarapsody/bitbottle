// Package perms implements the `bitbottle workspace perms` command group.
// Workspace permissions are a Bitbucket Cloud concept; the optional
// WorkspacePermsClient interface gates these commands.
package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdWorkspacePerms returns the `perms` sub-group command.
func NewCmdWorkspacePerms(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perms",
		Short: "Manage workspace permissions (Cloud only)",
	}
	cmd.AddCommand(NewCmdWorkspacePermsList(f, nil))
	cmd.AddCommand(NewCmdWorkspacePermsRepo(f))
	cmd.AddCommand(NewCmdWorkspacePermsGrant(f, nil))
	cmd.AddCommand(NewCmdWorkspacePermsRevoke(f, nil))
	return cmd
}
