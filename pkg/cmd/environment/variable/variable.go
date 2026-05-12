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
		Use:   "variable",
		Short: "Manage deployment environment variables",
	}
	cmd.AddCommand(list.NewCmdList(f, nil))
	cmd.AddCommand(set.NewCmdSet(f, nil))
	cmd.AddCommand(delete.NewCmdDelete(f, nil))
	return cmd
}
