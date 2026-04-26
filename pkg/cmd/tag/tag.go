package tag

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdTag(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}
	cmd.AddCommand(NewCmdTagList(f))
	cmd.AddCommand(NewCmdTagCreate(f))
	cmd.AddCommand(NewCmdTagDelete(f))
	return cmd
}
