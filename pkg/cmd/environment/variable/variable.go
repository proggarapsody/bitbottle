// Package variable wires the `environment variable` command group.
//
// Deprecated: use `bitbottle variable --scope deployment --env ENV-UUID` instead.
package variable

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdVariable builds the `environment variable` subcommand group.
func NewCmdVariable(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "variable",
		Short:      "Manage deployment environment variables",
		Deprecated: "use `bitbottle variable --scope deployment --env ENV-UUID` instead",
	}
	cmd.AddCommand(list.NewCmdList(f, nil))     //nolint:staticcheck
	cmd.AddCommand(set.NewCmdSet(f, nil))       //nolint:staticcheck
	cmd.AddCommand(delete.NewCmdDelete(f, nil)) //nolint:staticcheck
	return cmd
}
