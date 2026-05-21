package auth

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
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

	// 4+5. Single backend call covers both reachability and auth.
	// GetCurrentUser acts as the probe: transport errors → unreachable,
	// ErrAuth → reachable but token invalid, nil → fully authenticated.
	client, clientErr := f.Backend(host)
	if clientErr != nil {
		return fmt.Errorf("auth doctor: failed to build backend client: %w", clientErr)
	}

	if tokenStored {
		reachCheck, authCheck, passed := probeViaGetCurrentUser(client)
		if !passed {
			allPassed = false
		}
		checks = append(checks, reachCheck, authCheck)
	} else {
		// No token — still attempt a reachability probe (GetCurrentUser will
		// return ErrAuth since the backend uses a placeholder token, which still
		// means the host is reachable).
		reachCheck, _, _ := probeViaGetCurrentUser(client)
		checks = append(checks, reachCheck)
		checks = append(checks, doctorCheck{
			label:  "authenticated",
			passed: false,
			detail: "(skipped — no token)",
		})
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

// probeViaGetCurrentUser calls GetCurrentUser on the backend client and
// returns (reachCheck, authCheck, allPassed). It interprets the result as:
//   - nil error       → reachable: yes, authenticated: yes
//   - ErrTransport    → reachable: no, authenticated: unknown
//   - ErrAuth         → reachable: yes, authenticated: no (auth error)
//   - other DomainErr → reachable: yes, authenticated: no (<msg>)
//   - plain error     → reachable: unknown, authenticated: no
func probeViaGetCurrentUser(client backend.UserGetter) (reachCheck, authCheck doctorCheck, allPassed bool) {
	_, err := client.GetCurrentUser()
	if err == nil {
		return doctorCheck{label: "reachable", passed: true, detail: "yes"},
			doctorCheck{label: "authenticated", passed: true, detail: "yes"},
			true
	}

	var de *backend.DomainError
	if errors.As(err, &de) {
		switch {
		case errors.Is(de.Kind, backend.ErrTransport):
			return doctorCheck{label: "reachable", passed: false, detail: fmt.Sprintf("no (%s)", de.Error())},
				doctorCheck{label: "authenticated", passed: false, detail: "unknown"},
				false
		case errors.Is(de.Kind, backend.ErrAuth):
			return doctorCheck{label: "reachable", passed: true, detail: "yes"},
				doctorCheck{label: "authenticated", passed: false, detail: "no (auth error)"},
				false
		default:
			return doctorCheck{label: "reachable", passed: true, detail: "yes"},
				doctorCheck{label: "authenticated", passed: false, detail: fmt.Sprintf("no (%s)", de.Error())},
				false
		}
	}

	// Plain (non-domain) error — transport-level failure with no DomainError wrapping.
	return doctorCheck{label: "reachable", passed: false, detail: fmt.Sprintf("unknown (%s)", err.Error())},
		doctorCheck{label: "authenticated", passed: false, detail: "no"},
		false
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
