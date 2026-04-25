package auth

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuthLogin(f *factory.Factory) *cobra.Command {
	var hostname, gitProtocol string
	var skipTLS, withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hostname == "" {
				return fmt.Errorf("--hostname is required")
			}

			// Read token from stdin when --with-token is set.
			var token string
			if withToken {
				scanner := bufio.NewScanner(f.IOStreams.In)
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if token == "" {
					return fmt.Errorf("no token provided on stdin")
				}
			}

			// Validate the token by calling GetCurrentUser.
			client, err := f.Backend(hostname)
			if err != nil {
				return err
			}
			user, err := client.GetCurrentUser()
			if err != nil {
				return err
			}

			// Persist credentials to config.
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			cfg.Set(hostname, config.HostConfig{
				User:          user.Slug,
				OAuthToken:    token,
				GitProtocol:   gitProtocol,
				SkipTLSVerify: skipTLS,
			})
			if err := cfg.Save(); err != nil {
				return err
			}

			// Best-effort keyring storage — log error but do not fail.
			if krErr := f.Keyring.Set("bitbottle", user.Slug, token); krErr != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not store token in keyring: %v\n", krErr)
			}

			fmt.Fprintf(f.IOStreams.Out, "Logged in to %s as %s\n", hostname, user.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&gitProtocol, "git-protocol", "ssh", "Git protocol (ssh or https)")
	cmd.Flags().BoolVar(&skipTLS, "skip-tls-verify", false, "Skip TLS certificate verification")
	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")
	return cmd
}
