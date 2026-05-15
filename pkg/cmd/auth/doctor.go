package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// doctorCheck is the result of one diagnostic step.
type doctorCheck struct {
	label  string
	passed bool
	detail string
}

func (c doctorCheck) String() string {
	icon := "✓"
	if !c.passed {
		icon = "✗"
	}
	if c.detail != "" {
		return fmt.Sprintf("%s %s: %s", icon, c.label, c.detail)
	}
	return fmt.Sprintf("%s %s", icon, c.label)
}

// NewCmdDoctor returns the `auth doctor` command.
func NewCmdDoctor(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose credential and connectivity issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(f, hostname)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname to diagnose")
	return cmd
}

func runDoctor(f *factory.Factory, hostnameFlag string) error {
	cfg, host, err := resolveAuthHostname(f, hostnameFlag)
	if err != nil {
		return err
	}

	hostCfg, _ := cfg.Get(host)

	var checks []doctorCheck
	allPassed := true

	// 1. Keyring backend
	backendName := keyringBackendName()
	checks = append(checks, doctorCheck{
		label:  "keyring backend",
		passed: true,
		detail: backendName,
	})

	// 2. Token stored in keyring?
	token, keyringErr := f.Keyring.Get("bitbottle", hostCfg.User)
	tokenStored := keyringErr == nil && token != ""
	tokenCheck := doctorCheck{
		label:  "token stored",
		passed: tokenStored,
	}
	if tokenStored {
		tokenCheck.detail = "yes"
	} else {
		tokenCheck.detail = "no"
		allPassed = false
	}
	checks = append(checks, tokenCheck)

	// 3. Token format heuristic (only when token was found)
	if tokenStored {
		format := tokenFormat(token)
		checks = append(checks, doctorCheck{
			label:  "token format",
			passed: true,
			detail: format,
		})
	}

	// 4. Connectivity — unauthenticated GET to the API base URL
	isCloud := bbinstance.IsCloud(host, hostCfg.BackendType)
	var apiBase string
	if isCloud {
		apiBase = bbinstance.CloudRESTBase()
	} else {
		apiBase = bbinstance.RESTBase(host)
	}

	hc, err := f.HTTPClient(host)
	if err != nil {
		return fmt.Errorf("auth doctor: failed to build HTTP client: %w", err)
	}

	reachable, reachDetail := checkReachability(hc, apiBase)
	reachCheck := doctorCheck{
		label:  "reachable",
		passed: reachable,
		detail: reachDetail,
	}
	if !reachable {
		allPassed = false
	}
	checks = append(checks, reachCheck)

	// 5. Auth round-trip — authenticated GET /user or GET /rest/api/1.0/users/{slug}
	if tokenStored {
		authed, authDetail := checkAuthRoundTrip(hc, host, hostCfg.User, token, isCloud)
		authCheck := doctorCheck{
			label:  "authenticated",
			passed: authed,
			detail: authDetail,
		}
		if !authed {
			allPassed = false
		}
		checks = append(checks, authCheck)
	}

	// Output
	fmt.Fprintf(f.IOStreams.Out, "Diagnostics for %s\n\n", host)
	for _, c := range checks {
		fmt.Fprintln(f.IOStreams.Out, c.String())
	}

	if !allPassed {
		fmt.Fprintln(f.IOStreams.Out, "\n⚠ One or more checks failed. Run `bitbottle auth login` to re-authenticate.")
		return fmt.Errorf("auth doctor: one or more checks failed")
	}
	return nil
}

// keyringBackendName returns a human-readable label for the active keyring backend.
func keyringBackendName() string {
	if os.Getenv("BITBOTTLE_ALLOW_INSECURE_STORE") == "1" {
		return "file (insecure fallback)"
	}
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "windows":
		return "Windows Credential Manager"
	default:
		return "Secret Service (libsecret / GNOME Keyring)"
	}
}

// tokenFormat returns a human-readable label for the token without echoing the value.
func tokenFormat(token string) string {
	switch {
	case strings.HasPrefix(token, "BBDC-"):
		return "Server app-password (BBDC- prefix)"
	case strings.HasPrefix(token, "ATATT"):
		return "Cloud OAuth / Atlassian token (ATATT prefix)"
	default:
		return "unknown / app-password"
	}
}

// checkReachability performs an unauthenticated GET to baseURL and returns
// whether it succeeded along with a human-readable detail string.
func checkReachability(client factory.HTTPClient, baseURL string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return false, fmt.Sprintf("failed to build request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("unreachable: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	// Any HTTP response means the network layer works (even 401/403/404).
	return true, fmt.Sprintf("yes (HTTP %d)", resp.StatusCode)
}

// checkAuthRoundTrip calls the current-user endpoint with the stored token and
// reports whether authentication succeeded.
func checkAuthRoundTrip(client factory.HTTPClient, host, username, token string, isCloud bool) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var endpoint string
	var useBasicAuth bool
	if isCloud {
		endpoint = bbinstance.CloudRESTBase() + "/user"
		useBasicAuth = false
	} else {
		endpoint = bbinstance.RESTBase(host) + "/users/" + username
		useBasicAuth = true
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Sprintf("failed to build request: %v", err)
	}

	if useBasicAuth {
		req.SetBasicAuth(username, token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusOK {
		return true, "yes"
	}
	return false, fmt.Sprintf("no (HTTP %d)", resp.StatusCode)
}
