// Adapted from cli/cli (MIT) — pkg/cmdutil/repo_override.go.

package factory

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

// EnableRepoOverride registers a persistent -R/--repo flag on cmd and wires
// a PersistentPreRunE that swaps f.BaseRepo with an override resolver when
// the flag (or BB_REPO env) is set. The expected format is
// [HOST/]PROJECT/REPO; bare PROJECT/REPO uses the single configured host.
func EnableRepoOverride(cmd *cobra.Command, f *Factory) {
	cmd.PersistentFlags().StringP("repo", "R", "",
		"Select another repository using the `[HOST/]PROJECT/REPO` format")

	original := f.BaseRepo

	prev := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		if prev != nil {
			if err := prev(c, args); err != nil {
				return err
			}
		}
		repo, _ := c.Flags().GetString("repo")
		if repo == "" {
			repo = os.Getenv("BB_REPO")
		}
		if repo == "" {
			f.BaseRepo = original
			return nil
		}
		f.BaseRepo = func() (bbrepo.RepoRef, error) {
			return parseRepoOverride(repo, f)
		}
		return nil
	}
}

// parseRepoOverride parses [HOST/]PROJECT/REPO. When the host is omitted,
// fall back to the single configured host or error if multiple are present.
func parseRepoOverride(s string, f *Factory) (bbrepo.RepoRef, error) {
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 2:
		ref := bbrepo.RepoRef{Project: parts[0], Slug: parts[1]}
		cfg, err := f.Config()
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return bbrepo.RepoRef{}, fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			ref.Host = hosts[0]
		default:
			return bbrepo.RepoRef{}, fmt.Errorf("multiple hosts configured; specify host as HOST/PROJECT/REPO")
		}
		return ref, nil
	case 3:
		return bbrepo.RepoRef{Host: parts[0], Project: parts[1], Slug: parts[2]}, nil
	default:
		return bbrepo.RepoRef{}, fmt.Errorf("invalid --repo %q: expected [HOST/]PROJECT/REPO", s)
	}
}
