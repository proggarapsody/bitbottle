package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRList(f *factory.Factory) *cobra.Command {
	var state string
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pull requests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := resolveRepoRef(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prs, err := client.ListPRs(ref.Project, ref.Slug, strings.ToUpper(state), limit)
			if err != nil {
				return err
			}

			p := prFields(f, jsonFields, jqExpr)
			for _, pr := range prs {
				p.AddItem(pr)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "State filter: open, closed, merged")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of pull requests")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

// resolveRepoRef returns a RepoRef from an explicit argument or git remote detection.
// hostnameFlag overrides the host in all cases if non-empty.
func resolveRepoRef(f *factory.Factory, args []string, hostnameFlag string) (bbrepo.RepoRef, error) {
	var ref bbrepo.RepoRef
	var err error

	if len(args) == 1 {
		ref, err = bbrepo.Parse(args[0])
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
	} else {
		// No arg — detect from git remote.
		g := git.New(f.GitRunner())
		remoteURL, rerr := g.RemoteURL("origin")
		if rerr != nil {
			return bbrepo.RepoRef{}, fmt.Errorf("could not detect repo: %w; pass PROJECT/REPO as an argument", rerr)
		}
		ref, err = bbrepo.InferFromRemote(remoteURL)
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
	}

	// Explicit --hostname flag overrides detected host.
	if hostnameFlag != "" {
		ref.Host = hostnameFlag
		return ref, nil
	}

	// Host unknown (from PROJECT/REPO arg without host component): use config.
	if ref.Host == "" {
		cfg, cerr := f.Config()
		if cerr != nil {
			return bbrepo.RepoRef{}, cerr
		}
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return bbrepo.RepoRef{}, fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			ref.Host = hosts[0]
		default:
			return bbrepo.RepoRef{}, fmt.Errorf("multiple hosts configured; use --hostname to specify one")
		}
	}
	return ref, nil
}
