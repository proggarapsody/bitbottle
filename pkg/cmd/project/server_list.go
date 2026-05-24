package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdServerList builds `project server-list [--filter PREFIX] [--limit N] [--hostname HOST]`.
func NewCmdServerList(f *factory.Factory) *cobra.Command {
	var hostname string
	var filter string
	var limit int

	cmd := &cobra.Command{
		Use:   "server-list",
		Short: "List projects on a Bitbucket Server instance",
		Long: `List projects on a Bitbucket Server / Data Center host.

Use --filter to restrict to projects whose name starts with a prefix.
Bitbucket Cloud returns an unsupported error.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			pc, err := backend.AsServerProjectClient(client, host)
			if err != nil {
				return err
			}

			projects, err := pc.ListServerProjects(filter, limit)
			if err != nil {
				return err
			}

			p := serverProjectPrinter(f, format.ConfigFromCmd(cmd))
			for _, sp := range projects {
				p.AddItem(sp)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter projects by name prefix")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of projects to return")
	format.RegisterOutputFlags(cmd)
	return cmd
}
