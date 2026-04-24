package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
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
