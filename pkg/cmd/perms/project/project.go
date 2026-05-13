// Package project is the `perms project` subcommand tree.
package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/grant"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/revoke"
)

// NewCmdProject builds the `perms project` command tree.
func NewCmdProject(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project permissions",
	}
	cmd.AddCommand(list.NewCmdList(f, nil))
	cmd.AddCommand(grant.NewCmdGrant(f, nil))
	cmd.AddCommand(revoke.NewCmdRevoke(f, nil))
	return cmd
}
