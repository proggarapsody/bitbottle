package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRApprove(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve PR_ID",
		Short: "Approve a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	return cmd
}
