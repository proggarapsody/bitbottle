package auth

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuth(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with a Bitbucket host",
	}
	cmd.AddCommand(NewCmdAuthLogin(f))
	cmd.AddCommand(NewCmdAuthLogout(f))
	cmd.AddCommand(NewCmdAuthStatus(f))
	return cmd
}
