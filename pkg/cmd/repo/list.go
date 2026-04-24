package repo

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/internal/tableprinter"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdRepoList(f *factory.Factory) *cobra.Command {
	var limit int
	var jsonFields string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.HttpClient(host)
			if err != nil {
				return err
			}

			repos, err := client.ListRepos(limit)
			if err != nil {
				return err
			}

			if jsonFields != "" {
				return fmt.Errorf("--json not yet implemented")
			}

			if len(repos) == 0 {
				return nil
			}

			tp := tableprinter.New(f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), 80)
			tp.AddHeader("SLUG", "PROJECT", "TYPE")
			for _, r := range repos {
				tp.AddField(r.Slug)
				tp.AddField(r.Project.Key)
				tp.AddField(r.ScmID)
				tp.EndRow()
			}
			return tp.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of repositories")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output as JSON with specified fields")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (defaults to configured host)")
	return cmd
}

// resolveHostname returns the hostname to use: explicit flag, or the single
// configured host, or an error if ambiguous.
func resolveHostname(f *factory.Factory, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return hosts[0], nil
	default:
		return "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
