package group

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdList builds `group list [--filter PREFIX] [--hostname HOST]`.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var hostname string
	var filter string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List admin groups on Bitbucket Server/DC",
		Long: `List Bitbucket Server/DC admin groups.

Use --filter to restrict to groups whose name starts with a prefix.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			gc, err := backend.AsGroupClient(client, host)
			if err != nil {
				return err
			}

			groups, err := gc.ListGroups(filter, limit)
			if err != nil {
				return err
			}

			p := groupPrinter(f, format.ConfigFromCmd(cmd))
			for _, g := range groups {
				p.AddItem(g)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter groups by name prefix")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of groups to return")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func groupPrinter(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Group] {
	p := format.New[backend.Group](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Group]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(g backend.Group) any { return g.Name },
	})
	return p
}
