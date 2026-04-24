package pr

import (
	"fmt"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdPRCheckout(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout PR_ID",
		Short: "Check out a pull request branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	return cmd
}
