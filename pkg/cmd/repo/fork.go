package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoForkCreate(f *factory.Factory) *cobra.Command {
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
