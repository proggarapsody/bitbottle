// Package variable wires the `pipeline variable` command group.
//
// Deprecated: use `bitbottle variable --scope repository` instead.
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
		Use:        "variable",
		Short:      "Manage repository-level pipeline variables (Cloud only)",
		Deprecated: "use `bitbottle variable --scope repository` instead",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))    //nolint:staticcheck
	cmd.AddCommand(cmdSet.NewCmdSet(f, nil))     //nolint:staticcheck
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil)) //nolint:staticcheck
	return cmd
}
