// Package group implements the `group` command group for managing
// Bitbucket Server/DC admin groups and their members.
package group

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/group/member"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdGroup)
}

// NewCmdGroup builds the `group` top-level command.
func NewCmdGroup(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage Bitbucket Server/DC admin groups",
		Long: `Manage Bitbucket Server/DC admin groups and their members.

All group commands require a Bitbucket Server or Data Center host.
Bitbucket Cloud returns an unsupported error.`,
	}
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(member.NewCmdMember(f))
	return cmd
}
