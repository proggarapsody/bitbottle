package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRDiff(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff PR_ID",
		Short: "Show a pull request diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	return cmd
}
