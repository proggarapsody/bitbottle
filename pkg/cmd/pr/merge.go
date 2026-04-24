package pr

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdPRMerge(f *factory.Factory) *cobra.Command {
	var merge, squash, deleteBranch bool

	cmd := &cobra.Command{
		Use:   "merge PR_ID",
		Short: "Merge a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge commit strategy")
	cmd.Flags().BoolVar(&squash, "squash", false, "Squash merge strategy")
	cmd.Flags().BoolVar(&deleteBranch, "delete-branch", false, "Delete source branch after merge")
	_ = merge
	_ = squash
	_ = deleteBranch
	return cmd
}
