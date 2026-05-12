// Package profile implements the `profile` command group for managing named
// credential profiles (kubectl-context-like).
package profile

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdCreate "github.com/proggarapsody/bitbottle/pkg/cmd/profile/create"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/profile/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/profile/list"
	cmdUse "github.com/proggarapsody/bitbottle/pkg/cmd/profile/use"
)

// NewCmdProfile builds the root `profile` command.
func NewCmdProfile(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage named credential profiles",
	}
	cmd.AddCommand(cmdCreate.NewCmdCreate(f, nil))
	cmd.AddCommand(cmdUse.NewCmdUse(f, nil))
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	return cmd
}
