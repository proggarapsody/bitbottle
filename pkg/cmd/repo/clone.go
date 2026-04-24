package repo

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdRepoClone(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone PROJECT/REPO [DIR]",
		Short: "Clone a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	return cmd
}
