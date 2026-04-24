package repo

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdRepoDelete(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO]",
		Short: "Delete a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	_ = confirm
	return cmd
}
