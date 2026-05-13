// Package repo is the `perms repo` subcommand tree.
package repo

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/grant"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/revoke"
)

// NewCmdRepo builds the `perms repo` command tree.
func NewCmdRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repository permissions",
	}
	cmd.AddCommand(list.NewCmdList(f, nil))
	cmd.AddCommand(grant.NewCmdGrant(f, nil))
	cmd.AddCommand(revoke.NewCmdRevoke(f, nil))
	return cmd
}
