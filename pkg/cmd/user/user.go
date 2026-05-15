// Package user implements the `bitbottle user` command group.
package user

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/user/view"
)

func NewCmdUser(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "View and manage user information",
	}
	cmd.AddCommand(cmdView.NewCmdView(f, nil))
	return cmd
}
