package branch

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage branches",
	}
	cmd.AddCommand(NewCmdBranchList(f))
	cmd.AddCommand(NewCmdBranchDelete(f))
	cmd.AddCommand(NewCmdBranchCreate(f))
	cmd.AddCommand(NewCmdBranchCheckout(f))
	return cmd
}
