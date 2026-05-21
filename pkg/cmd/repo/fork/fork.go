package fork

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdForkCreate(f *factory.Factory) *cobra.Command {
	var hostname, into, name string

	cmd := &cobra.Command{
		Use:   "create PROJECT/REPO --into WORKSPACE [--name NAME]",
		Short: "Fork a repository into a workspace (Bitbucket Cloud only)",
		Long: "Fork a Bitbucket Cloud repository into a destination workspace.\n" +
			"Bitbucket Server / Data Center has no fork primitive — running this\n" +
			"against a Server host returns a typed unsupported-capability error.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			forker, err := backend.AsRepoForker(client, host)
			if err != nil {
				return err
			}

			fork, err := forker.ForkRepo(ref.Project, ref.Slug, backend.ForkRepoInput{
				Workspace: into,
				Name:      name,
			})
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(fork)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Forked %s/%s to %s/%s\n",
				ref.Project, ref.Slug, fork.Namespace, fork.Slug)
			if fork.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", fork.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&into, "into", "", "Destination workspace slug (required)")
	cmd.Flags().StringVar(&name, "name", "", "Override the fork's name (defaults to source name)")
	_ = cmd.MarkFlagRequired("into")
	return cmd
}

func repoFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Repository] {
	p := format.New[backend.Repository](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Repository]{Name: "slug", Header: "SLUG", Extract: func(r backend.Repository) any { return r.Slug }})
	p.AddField(format.Field[backend.Repository]{Name: "name", Header: "NAME", Extract: func(r backend.Repository) any { return r.Name }})
	p.AddField(format.Field[backend.Repository]{Name: "namespace", Header: "PROJECT", Extract: func(r backend.Repository) any { return r.Namespace }})
	p.AddField(format.Field[backend.Repository]{Name: "scm", Header: "TYPE", Extract: func(r backend.Repository) any { return r.SCM }})
	p.AddField(format.Field[backend.Repository]{Name: "description", Header: "DESCRIPTION", Extract: func(r backend.Repository) any { return r.Description }})
	p.AddField(format.Field[backend.Repository]{Name: "webURL", Header: "URL", Extract: func(r backend.Repository) any { return r.WebURL }})
	return p
}

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
