// Package variable wires the top-level `variable` command group.
package variable

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/variable/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/variable/list"
	cmdSet "github.com/proggarapsody/bitbottle/pkg/cmd/variable/set"
)

// NewCmdVariable builds the root `variable` command.
func NewCmdVariable(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage repository, workspace, and deployment variables",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdSet.NewCmdSet(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	return cmd
}
