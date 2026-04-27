package auth

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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

			// Shared scanner — reused across all interactive prompts so that
			// buffered reads on one prompt don't consume input meant for the next.
			scanner := bufio.NewScanner(f.IOStreams.In)

			// ------------------------------------------------------------------
			// 1. Collect token
			// ------------------------------------------------------------------
			var token string
			switch {
			case withToken:
				// Non-interactive: read a single line from stdin.
				if scanner.Scan() {
					token = strings.TrimSpace(scanner.Text())
				}
				if token == "" {
					return fmt.Errorf("no token provided on stdin")
				}

			case f.IOStreams.IsStdoutTTY():
				// Interactive guided flow.

				// For Server/DC, ask for the username first so we can
				// embed it in the PAT management URL.
				if !bbinstance.IsCloud(hostname, "") && username == "" {
					if cfg, err := f.Config(); err == nil {
						if h, ok := cfg.Get(hostname); ok {
							username = h.User
						}
					}
					if username == "" {
						fmt.Fprintf(f.IOStreams.Out, "Bitbucket username for %s: ", hostname)
						if scanner.Scan() {
							username = strings.TrimSpace(scanner.Text())
						}
						if username == "" {
							return fmt.Errorf("username is required for Bitbucket Server/Data Center instances")
						}
					}
				}

				// Build the PAT URL now that username is known; probe which format this instance uses.
				tokenURL := patURL(f, hostname, username, skipTLS)

				// Ask how the user wants to authenticate.
				fmt.Fprintf(f.IOStreams.Out, "\nHow would you like to authenticate with %s?\n", hostname)
				fmt.Fprintf(f.IOStreams.Out, "  1. Open browser to create a Personal Access Token\n")
				fmt.Fprintf(f.IOStreams.Out, "  2. Paste a Personal Access Token directly\n")
				fmt.Fprintf(f.IOStreams.Out, "\nChoice [1]: ")

				choice := "1"
				if scanner.Scan() {
					if s := strings.TrimSpace(scanner.Text()); s != "" {
						choice = s
					}
				}

				switch choice {
				case "1":
					fmt.Fprintf(f.IOStreams.Out, "\nOpening %s in your browser...\n", tokenURL)
					if browseErr := f.Browser.Browse(tokenURL); browseErr != nil {
						fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not open browser: %v\n", browseErr)
						fmt.Fprintf(f.IOStreams.Out, "Visit this URL manually:\n  %s\n", tokenURL)
					}
					fmt.Fprintf(f.IOStreams.Out, "Press Enter once you have your token.\n\n")
					scanner.Scan() // wait for Enter
				case "2":
					fmt.Fprintf(f.IOStreams.Out, "\nCreate a token at:\n  %s\n\n", tokenURL)
				default:
					return fmt.Errorf("invalid choice %q: enter 1 or 2", choice)
				}

				fmt.Fprintf(f.IOStreams.Out, "Paste your Personal Access Token: ")
				var err error
				token, err = readSecret(f.IOStreams, scanner)
				if err != nil {
					return fmt.Errorf("could not read token: %w", err)
				}
				if token == "" {
					return fmt.Errorf("no token entered")
				}

			default:
				// Non-TTY without --with-token: fall back to stored token.
				// Useful for re-validating a previously stored credential.
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

			// ------------------------------------------------------------------
			// 2. Collect username for Server/DC (non-TTY path or --with-token)
			// ------------------------------------------------------------------
			// The TTY path already asked above; this block handles the remaining
			// non-interactive cases where username wasn't provided yet.
			if !bbinstance.IsCloud(hostname, "") && username == "" {
				if cfg, err := f.Config(); err == nil {
					if h, ok := cfg.Get(hostname); ok {
						username = h.User
					}
				}
				if username == "" {
					return fmt.Errorf("--username is required for Bitbucket Server/Data Center instances")
				}
			}

			// ------------------------------------------------------------------
			// 3. Validate credentials against the API
			// ------------------------------------------------------------------
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

			// ------------------------------------------------------------------
			// 4. Persist credentials
			// ------------------------------------------------------------------
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

			fmt.Fprintf(f.IOStreams.Out, "\n✓ Logged in to %s as %s\n", hostname, user.Slug)
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

// patURL returns the Personal Access Token management URL for the given host.
// For Cloud it always returns the App Passwords page. For Server/DC it probes
// which URL format the running instance uses before returning.
func patURL(f *factory.Factory, hostname, username string, skipTLS bool) string {
	if bbinstance.IsCloud(hostname, "") {
		return bbinstance.CloudAppPasswordsURL()
	}
	prober := f.ServerPATURLProber
	if prober == nil {
		prober = probeServerPATURL
	}
	return prober(hostname, username, skipTLS)
}

// probeServerPATURL checks the two known Bitbucket Server PAT URL patterns
// and returns the first one the server acknowledges with a non-404 status.
// An unauthenticated HEAD request is sufficient: valid endpoints return 401,
// missing ones return 404. Falls back to user-scoped URL if probing fails.
func probeServerPATURL(hostname, username string, skipTLS bool) string {
	candidates := []string{
		bbinstance.PATManageURL(hostname, username), // user-scoped (older/some versions)
		bbinstance.PATManageURL(hostname, ""),       // generic   (newer versions)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS}, //nolint:gosec // mirrors --skip-tls-verify
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
		// Don't follow redirects — the raw status is what matters.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, u := range candidates {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodHead, u, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			return u
		}
	}

	return candidates[0] // best guess
}

// readSecret reads a secret from the terminal without echoing the input.
// When the underlying reader is an *os.File (real terminal), it uses
// term.ReadPassword to suppress echo. Otherwise (tests, piped input) it
// reads the next line from the shared fallback scanner to avoid buffering
// conflicts with callers that already hold a scanner over the same reader.
func readSecret(ios *iostreams.IOStreams, fallback *bufio.Scanner) (string, error) {
	if f, ok := ios.In.(*os.File); ok {
		raw, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(ios.Out) // emit newline after the hidden input
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(raw)), nil
	}
	// Fallback: reuse the caller's scanner so buffered bytes are not lost.
	if fallback.Scan() {
		return strings.TrimSpace(fallback.Text()), nil
	}
	return "", nil
}
