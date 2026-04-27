package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func NewCmdAuthLogin(f *factory.Factory) *cobra.Command {
	var hostname, gitProtocol, username string
	var skipTLS, withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hostname == "" {
				return fmt.Errorf("--hostname is required")
			}

			// Collect token: from stdin (--with-token), interactive prompt, or
			// fall back to whatever is already stored in the config file.
			var token string
			switch {
			case withToken:
				scanner := bufio.NewScanner(f.IOStreams.In)
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if token == "" {
					return fmt.Errorf("no token provided on stdin")
				}
			case f.IOStreams.IsStdoutTTY():
				fmt.Fprintf(f.IOStreams.Out, "Paste your Personal Access Token for %s: ", hostname)
				var err error
				token, err = readSecret(f.IOStreams)
				if err != nil {
					return fmt.Errorf("could not read token: %w", err)
				}
				if token == "" {
					return fmt.Errorf("no token entered")
				}
			default:
				// Non-TTY without --with-token: read from existing config.
				// Useful for re-validating a previously stored token.
				cfg, err := f.Config()
				if err != nil {
					return err
				}
				if h, ok := cfg.Get(hostname); ok {
					token = h.OAuthToken
				}
				if token == "" {
					return fmt.Errorf("no token: use --with-token to pass a PAT on stdin")
				}
			}

			// Bitbucket Server/Data Center does not support GET /users/~ as a
			// self-reference; a username is needed to call GET /users/{slug}.
			// Resolution order: --username flag → stored config → TTY prompt → error.
			if !bbinstance.IsCloud(hostname, "") && username == "" {
				if cfg, err := f.Config(); err == nil {
					if h, ok := cfg.Get(hostname); ok {
						username = h.User
					}
				}
				if username == "" && f.IOStreams.IsStdoutTTY() {
					fmt.Fprintf(f.IOStreams.Out, "Bitbucket username for %s: ", hostname)
					scanner := bufio.NewScanner(f.IOStreams.In)
					if scanner.Scan() {
						username = strings.TrimSpace(scanner.Text())
					}
				}
				if username == "" {
					return fmt.Errorf("--username is required for Bitbucket Server/Data Center instances")
				}
			}

			// Build a one-shot HTTP client that honours --skip-tls-verify
			// immediately, before any token is stored in the config file.
			client, err := f.BackendWithOptions(hostname, backend.Options{
				Token:         token,
				SkipTLSVerify: skipTLS,
				Username:      username,
			})
			if err != nil {
				return err
			}
			user, err := client.GetCurrentUser()
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Persist credentials.
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

			// Best-effort keyring storage.
			if krErr := f.Keyring.Set("bitbottle", user.Slug, token); krErr != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not store token in keyring: %v\n", krErr)
			}

			fmt.Fprintf(f.IOStreams.Out, "Logged in to %s as %s\n", hostname, user.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&gitProtocol, "git-protocol", "ssh", "Git protocol (ssh or https)")
	cmd.Flags().StringVar(&username, "username", "", "Bitbucket username (required for Server/Data Center)")
	cmd.Flags().BoolVar(&skipTLS, "skip-tls-verify", false, "Skip TLS certificate verification")
	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")
	return cmd
}

// readSecret reads a secret from the terminal without echoing the input.
// When the underlying reader is an *os.File (real terminal), it uses
// term.ReadPassword to suppress echo. Otherwise (tests, piped input) it
// falls back to a plain bufio.Scanner line read.
func readSecret(ios *iostreams.IOStreams) (string, error) {
	if f, ok := ios.In.(*os.File); ok {
		raw, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(ios.Out) // emit newline after the hidden input
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(raw)), nil
	}
	// Fallback: plain readline — used in tests and when stdin is a buffer.
	scanner := bufio.NewScanner(ios.In)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", nil
}
