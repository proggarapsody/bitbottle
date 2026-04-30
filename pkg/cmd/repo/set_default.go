package repo

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Local git config keys consulted by the factory's BaseRepo before falling
// back to git remote inference. Documented here as the source of truth for
// other code that reads them.
const (
	GitConfigHostKey    = "bitbottle.host"
	GitConfigProjectKey = "bitbottle.project"
	GitConfigSlugKey    = "bitbottle.slug"
)

// NewCmdRepoSetDefault implements `bitbottle repo set-default
// [HOST/]PROJECT/REPO`. The chosen coordinate is pinned in the local
// repository's git config so subsequent commands do not infer a different
// repo from the git remote — the seamless wedge for daily users and AI
// agents working in a checkout that may have multiple remotes or none.
func NewCmdRepoSetDefault(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default [HOST/]PROJECT/REPO",
		Short: "Pin the default Bitbucket repository for this checkout",
		Long: "Writes bitbottle.host, bitbottle.project, and bitbottle.slug into\n" +
			"the local repository's git config. Future commands in this checkout\n" +
			"resolve to that repository without consulting the git remote or\n" +
			"requiring -R / --hostname flags.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, project, slug, err := parseSetDefaultArg(args[0], f)
			if err != nil {
				return err
			}
			g := git.New(f.GitRunner())
			if err := g.SetConfig(GitConfigHostKey, host); err != nil {
				return fmt.Errorf("write %s: %w", GitConfigHostKey, err)
			}
			if err := g.SetConfig(GitConfigProjectKey, project); err != nil {
				return fmt.Errorf("write %s: %w", GitConfigProjectKey, err)
			}
			if err := g.SetConfig(GitConfigSlugKey, slug); err != nil {
				return fmt.Errorf("write %s: %w", GitConfigSlugKey, err)
			}
			fmt.Fprintf(f.IOStreams.Out, "Default repository set to %s/%s on %s\n", project, slug, host)
			return nil
		},
	}
	return cmd
}

// parseSetDefaultArg accepts `HOST/PROJECT/REPO` (3 parts) or `PROJECT/REPO`
// (2 parts; host inferred from the single configured host).
func parseSetDefaultArg(arg string, f *factory.Factory) (host, project, slug string, err error) {
	parts := strings.Split(arg, "/")
	switch len(parts) {
	case 2:
		project, slug = parts[0], parts[1]
		cfg, cerr := f.Config()
		if cerr != nil {
			return "", "", "", cerr
		}
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return "", "", "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			host = hosts[0]
		default:
			return "", "", "", fmt.Errorf("multiple hosts configured; specify host as HOST/PROJECT/REPO")
		}
		return host, project, slug, nil
	case 3:
		return parts[0], parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("invalid argument %q: expected [HOST/]PROJECT/REPO", arg)
	}
}
