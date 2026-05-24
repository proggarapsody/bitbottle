package pat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// scopeAliases maps user-friendly scope aliases to Bitbucket Server permission names.
var scopeAliases = map[string]string{
	"repo:read":     "REPO_READ",
	"repo:write":    "REPO_WRITE",
	"pr:read":       "PR_READ",
	"pr:write":      "PR_WRITE",
	"project:read":  "PROJECT_READ",
	"project:write": "PROJECT_WRITE",
}

// NewCmdCreate builds `auth pat create --name NAME --scopes S1,S2 [--expires-in DAYS] [--hostname H]`.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var hostname string
	var name string
	var scopes string
	var expiresIn int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a personal access token on Bitbucket Server/DC",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if scopes == "" {
				return fmt.Errorf("--scopes is required")
			}

			permissions, err := resolveScopes(scopes)
			if err != nil {
				return err
			}

			host, userSlug, err := resolveHostAndUser(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			pc, err := backend.AsPATClient(client, host)
			if err != nil {
				return err
			}

			in := backend.CreatePATInput{
				Name:        name,
				Permissions: permissions,
			}
			if cmd.Flags().Changed("expires-in") {
				in.ExpiryDays = &expiresIn
			}

			pat, err := pc.CreatePAT(userSlug, in)
			if err != nil {
				return err
			}

			fmt.Fprintln(f.IOStreams.Out, "⚠ Store this token now — it will not be shown again.")
			fmt.Fprintf(f.IOStreams.Out, "Token: %s\n", pat.Token)
			fmt.Fprintf(f.IOStreams.Out, "Created PAT %s (%s)\n", pat.ID, pat.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&name, "name", "", "Token name (required)")
	cmd.Flags().StringVar(&scopes, "scopes", "", "Comma-separated scopes: repo:read, repo:write, pr:read, pr:write, project:read, project:write (required)")
	cmd.Flags().IntVar(&expiresIn, "expires-in", 0, "Token lifetime in days (omit for no expiry)")
	return cmd
}

// resolveScopes converts a comma-separated scope string to Bitbucket permission names.
// Accepts both alias form (repo:read) and canonical form (REPO_READ).
func resolveScopes(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if canon, ok := scopeAliases[p]; ok {
			out = append(out, canon)
			continue
		}
		// Accept canonical names directly (e.g. REPO_READ).
		upper := strings.ToUpper(p)
		valid := false
		for _, v := range scopeAliases {
			if v == upper {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("unknown scope %q; valid scopes: repo:read, repo:write, pr:read, pr:write, project:read, project:write", p)
		}
		out = append(out, upper)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--scopes must not be empty")
	}
	return out, nil
}
