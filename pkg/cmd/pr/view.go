package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view PR_ID",
		Short: "View a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	_ = web
	return cmd
}
