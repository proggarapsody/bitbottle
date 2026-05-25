package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdWorkspacePermsRepo returns the `workspace perms repo` sub-group.
func NewCmdWorkspacePermsRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage workspace repo permissions (Cloud only)",
	}
	cmd.AddCommand(NewCmdWorkspacePermsRepoList(f, nil))
	return cmd
}
