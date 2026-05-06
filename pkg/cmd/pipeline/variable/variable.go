// Package variable wires the `pipeline variable` command group.
package variable

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/list"
	cmdSet "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/set"
)

// NewCmdVariable builds the root `pipeline variable` command.
func NewCmdVariable(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage repository-level pipeline variables (Cloud only)",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdSet.NewCmdSet(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	return cmd
}
