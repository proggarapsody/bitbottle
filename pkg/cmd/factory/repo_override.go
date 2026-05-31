// Adapted from cli/cli (MIT) — pkg/cmdutil/repo_override.go.

package factory

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// EnableRepoOverride registers a persistent -R/--repo flag on cmd and wires
// a PersistentPreRunE that swaps f.BaseRepo with an override resolver when
// the flag (or BB_REPO env) is set. The expected format is
// [HOST/]PROJECT/REPO; bare PROJECT/REPO uses the single configured host.
//
// Parsing/host-inference for the override is delegated to ResolveTarget so
// the rules stay in one place.
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
		// Cobra runs only the single deepest PersistentPreRunE in a chain, so
		// this hook shadows the root's output-format guard for commands under
		// this group. Re-apply the shared contract here so --jq/--json/--yaml/
		// --template rules and color-disable hold uniformly (FMT-CONTRACT).
		cmdutil.ApplyNoColorFlag(c, f.IOStreams)
		if err := cmdutil.ValidateOutputFlags(c, f.IOStreams); err != nil {
			return err
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
			// Delegate to the unified resolver. Pass empty hostnameFlag —
			// --hostname is applied at the command level, not here.
			return ResolveTarget(f, []string{repo}, "")
		}
		return nil
	}
}
