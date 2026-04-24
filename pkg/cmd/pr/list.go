package pr

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/tableprinter"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdPRList(f *factory.Factory) *cobra.Command {
	var state string
	var limit int
	var jsonFields string
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

			client, err := f.HttpClient(ref.Host)
			if err != nil {
				return err
			}

			prs, err := client.ListPRs(ref.Project, ref.Slug, strings.ToUpper(state), limit)
			if err != nil {
				return err
			}

			if jsonFields != "" {
				return fmt.Errorf("--json not yet implemented")
			}

			if len(prs) == 0 {
				return nil
			}

			tp := tableprinter.New(f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), 80)
			tp.AddHeader("TITLE", "AUTHOR", "STATE")
			for _, p := range prs {
				tp.AddField(p.Title)
				tp.AddField(p.Author.User.Slug)
				tp.AddField(p.State)
				tp.EndRow()
			}
			return tp.Render()
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "State filter: open, closed, merged")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of pull requests")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output as JSON with specified fields")
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
