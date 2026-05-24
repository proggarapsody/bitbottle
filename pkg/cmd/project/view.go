package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdView builds `project view KEY [--hostname HOST] [--json]`.
func NewCmdView(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "view KEY",
		Short: "View a project on Bitbucket Server",
		Long: `View details of a project on Bitbucket Server / Data Center.

Bitbucket Cloud returns an unsupported error.

Examples:
  bitbottle project view PRJ --hostname git.example.com
  bitbottle project view PRJ --hostname git.example.com --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

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

			p, err := pc.GetServerProject(key)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format == format.FormatJSON {
				printer := serverProjectPrinter(f, cfg)
				printer.AddItem(p)
				return printer.Render()
			}

			// Human-readable output
			fmt.Fprintf(f.IOStreams.Out, "Key:         %s\n", p.Key)
			fmt.Fprintf(f.IOStreams.Out, "Name:        %s\n", p.Name)
			fmt.Fprintf(f.IOStreams.Out, "Description: %s\n", p.Description)
			fmt.Fprintf(f.IOStreams.Out, "Public:      %v\n", p.Public)
			fmt.Fprintf(f.IOStreams.Out, "URL:         %s\n", p.WebURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func serverProjectPrinter(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.ServerProject] {
	p := format.New[backend.ServerProject](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.ServerProject]{Name: "key", Header: "KEY", Extract: func(sp backend.ServerProject) any { return sp.Key }})
	p.AddField(format.Field[backend.ServerProject]{Name: "name", Header: "NAME", Extract: func(sp backend.ServerProject) any { return sp.Name }})
	p.AddField(format.Field[backend.ServerProject]{Name: "description", Header: "DESCRIPTION", JSONOnly: true, Extract: func(sp backend.ServerProject) any { return sp.Description }})
	p.AddField(format.Field[backend.ServerProject]{Name: "public", Header: "PUBLIC", Extract: func(sp backend.ServerProject) any { return sp.Public }})
	p.AddField(format.Field[backend.ServerProject]{Name: "webURL", Header: "URL", Aliases: []string{"url"}, Extract: func(sp backend.ServerProject) any { return sp.WebURL }})
	return p
}
