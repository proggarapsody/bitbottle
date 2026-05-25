// Package perms implements the `bitbottle workspace project perms` command group.
// Workspace project permissions are a Bitbucket Cloud concept; the optional
// WorkspaceProjectPermsClient interface gates these commands.
package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdWorkspaceProjectPerms returns the `perms` sub-group under
// `workspace project`.
func NewCmdWorkspaceProjectPerms(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perms",
		Short: "Manage workspace project permissions (Cloud only)",
	}
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdGrant(f, nil))
	cmd.AddCommand(NewCmdRevoke(f, nil))
	return cmd
}
