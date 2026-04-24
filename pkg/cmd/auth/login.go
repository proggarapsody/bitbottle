package auth

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdAuthLogin(f *factory.Factory) *cobra.Command {
	var hostname, gitProtocol string
	var skipTLS, withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&gitProtocol, "git-protocol", "ssh", "Git protocol (ssh or https)")
	cmd.Flags().BoolVar(&skipTLS, "skip-tls-verify", false, "Skip TLS certificate verification")
	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")
	_ = hostname
	_ = gitProtocol
	_ = skipTLS
	_ = withToken
	return cmd
}
