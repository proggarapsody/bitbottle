package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO]",
		Short: "View a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	_ = web
	return cmd
}
