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
	"github.com/proggarapsody/bitbottle/internal/tlsprobe"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func NewCmdAuthLogin(f *factory.Factory) *cobra.Command {
	var hostname, gitProtocol, username, email string
	var skipTLS, withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hostname == "" {
				return fmt.Errorf("--hostname is required")
			}

			// Strip any http:// or https:// prefix and trailing slash so
			// downstream URL builders (bbinstance.RESTBase, PATManageURL,
			// HTTPSURL) cannot produce double-scheme URLs like
			// "https://https://HOST/...". See PRD #372 Bug A.
			hostname = normalizeHostname(hostname)

			// Flag cross-validation: --username is for Server/DC only;
			// --email is for Cloud only.
			isCloud := bbinstance.IsCloud(hostname, "")
			if isCloud && username != "" {
				return fmt.Errorf("--username is not supported for Bitbucket Cloud; use --email (your Atlassian account email) instead")
			}
			if !isCloud && email != "" {
				return fmt.Errorf("--email is not supported for Bitbucket Server/Data Center; use --username instead")
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

				if isCloud {
					// Cloud: ask for the Atlassian account email so we can use
					// Basic auth with the API token.
					if email == "" {
						if cfg, err := f.Config(); err == nil {
							if h, ok := cfg.Get(hostname); ok {
								email = h.AuthUser
							}
						}
					}
					if email == "" {
						fmt.Fprintf(f.IOStreams.Out, "Atlassian account email for %s: ", hostname)
						if scanner.Scan() {
							email = strings.TrimSpace(scanner.Text())
						}
						if email == "" {
							return fmt.Errorf("email is required for Bitbucket Cloud (use --email)")
						}
					}
				} else {
					// Server/DC: ask for the Bitbucket username so we can
					// embed it in the PAT management URL.
					if username == "" {
						if cfg, err := f.Config(); err == nil {
							if h, ok := cfg.Get(hostname); ok {
								username = h.User
							}
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
					scanner.Scan()
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

			if isCloud {
				if email == "" {
					if cfg, err := f.Config(); err == nil {
						if h, ok := cfg.Get(hostname); ok {
							email = h.AuthUser
						}
					}
				}
				if email == "" {
					return fmt.Errorf("--email is required for Bitbucket Cloud (your Atlassian account email)")
				}
			} else {
				if username == "" {
					if cfg, err := f.Config(); err == nil {
						if h, ok := cfg.Get(hostname); ok {
							username = h.User
						}
					}
				}
				if username == "" {
					return fmt.Errorf("--username is required for Bitbucket Server/Data Center instances")
				}
			}

			// TLS auto-trust probe (Server/DC only, and only when the
			// user did not already pass --skip-tls-verify). If the
			// host's cert chains to an OS-trusted CA the probe is a
			// no-op; otherwise we show the cert and let the user
			// trust it interactively — same UX as SSH known-hosts.
			if !isCloud && !skipTLS && f.TLSProber != nil {
				probeCtx, probeCancel := context.WithTimeout(cmd.Context(), 10*time.Second)
				res, perr := f.TLSProber(probeCtx, hostname+":443", tlsprobe.Options{})
				probeCancel()
				switch {
				case perr != nil:
					// Network/DNS errors are not the probe's
					// problem — keep going so the backend call
					// surfaces the real error from the catalogue.
					fmt.Fprintf(f.IOStreams.ErrOut, "warning: TLS pre-flight probe failed: %v\n", perr)
				case res != nil && !res.TrustedByOS:
					trusted, terr := confirmSelfSignedCert(f.IOStreams, scanner, hostname, res)
					if terr != nil {
						return terr
					}
					skipTLS = trusted
				}
			}

			client, err := f.BackendWithOptions(hostname, backend.Options{
				Token:         token,
				SkipTLSVerify: skipTLS,
				Email:         email,
				Username:      username,
			})
			if err != nil {
				return err
			}
			user, err := client.GetCurrentUser()
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			authUser := email
			if authUser == "" {
				authUser = username
			}
			cfg.Set(hostname, config.HostConfig{
				User:          user.Slug,
				AuthUser:      authUser,
				OAuthToken:    token,
				GitProtocol:   gitProtocol,
				SkipTLSVerify: skipTLS,
			})
			if err := cfg.Save(); err != nil {
				return err
			}

			if krErr := f.Keyring.Set("bitbottle", user.Slug, token); krErr != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not store token in keyring: %v\n", krErr)
			}

			fmt.Fprintf(f.IOStreams.Out, "\n✓ Logged in to %s as %s\n", hostname, user.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&gitProtocol, "git-protocol", "ssh", "Git protocol (ssh or https)")
	cmd.Flags().StringVar(&email, "email", "", "Atlassian account email (required for Bitbucket Cloud API token auth)")
	cmd.Flags().StringVar(&username, "username", "", "Bitbucket username (required for Bitbucket Server/Data Center)")
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

// normalizeHostname strips a leading http:// or https:// scheme (case-
// insensitive) and a single trailing slash. Downstream URL builders in
// internal/bbinstance interpolate the hostname into "https://%s/..." —
// without this normalization, passing --hostname https://HOST produces
// "https://https://HOST/...", which Go's URL parser collapses to a DNS
// lookup of the literal hostname "https". See PRD #372 Bug A.
func normalizeHostname(s string) string {
	low := strings.ToLower(s)
	switch {
	case strings.HasPrefix(low, "https://"):
		s = s[len("https://"):]
	case strings.HasPrefix(low, "http://"):
		s = s[len("http://"):]
	}
	return strings.TrimSuffix(s, "/")
}

// confirmSelfSignedCert renders the leaf cert from a TLS probe result
// and asks the user whether to trust it. Returns (true, nil) when the
// user confirms — caller should set skip_tls_verify=true. Returns
// (false, error) when the user declines OR the environment is
// non-interactive (no TTY, or BB_PROMPT_DISABLED set) — in that case
// the error explains how to recover with --skip-tls-verify.
func confirmSelfSignedCert(ios *iostreams.IOStreams, scanner *bufio.Scanner, hostname string, res *tlsprobe.Result) (bool, error) {
	if !ios.IsStdoutTTY() || os.Getenv("BB_PROMPT_DISABLED") != "" {
		return false, fmt.Errorf(
			"certificate at %s is not trusted by the OS (issuer: %s, sha256: %s); re-run with --skip-tls-verify if you trust this host",
			hostname,
			res.LeafCert.Issuer.CommonName,
			res.FingerprintSHA256,
		)
	}

	fmt.Fprintln(ios.Out)
	fmt.Fprintf(ios.Out, "TLS verification failed for %s — the certificate is not signed by any CA your OS trusts.\n", hostname)
	fmt.Fprintf(ios.Out, "  Subject:     %s\n", res.LeafCert.Subject.CommonName)
	fmt.Fprintf(ios.Out, "  Issuer:      %s\n", res.LeafCert.Issuer.CommonName)
	if !res.LeafCert.NotBefore.IsZero() && !res.LeafCert.NotAfter.IsZero() {
		fmt.Fprintf(ios.Out, "  Valid:       %s → %s\n",
			res.LeafCert.NotBefore.UTC().Format("2006-01-02"),
			res.LeafCert.NotAfter.UTC().Format("2006-01-02"),
		)
	}
	fmt.Fprintf(ios.Out, "  SHA-256:     %s\n", res.FingerprintSHA256)
	fmt.Fprintf(ios.Out, "Trust this certificate? [y/N]: ")

	var answer string
	if scanner.Scan() {
		answer = strings.TrimSpace(scanner.Text())
	}
	if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
		return false, fmt.Errorf("certificate not trusted by user; aborting login (re-run with --skip-tls-verify to force, or install the CA in your OS trust store)")
	}
	fmt.Fprintf(ios.ErrOut, "Trusting self-signed cert — installing the corporate CA into your OS trust store is strictly safer.\n")
	return true, nil
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
