package label

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLabel is the parent of `repo label <action>`.
func NewCmdLabel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage repository labels",
	}
	cmd.AddCommand(NewCmdLabelList(f))
	cmd.AddCommand(NewCmdLabelCreate(f))
	cmd.AddCommand(NewCmdLabelUpdate(f))
	cmd.AddCommand(NewCmdLabelDelete(f))
	return cmd
}
