package repo

import (
	"fmt"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
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
