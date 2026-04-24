package auth

import (
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
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
