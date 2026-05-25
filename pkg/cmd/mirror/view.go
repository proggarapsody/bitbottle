package mirror

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ViewOptions carries parsed flags for `mirror view`.
type ViewOptions struct {
	Output   format.OutputConfig
	Hostname string
	ID       string
}

// NewCmdMirrorView constructs the `mirror view` cobra command.
func NewCmdMirrorView(f *factory.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{}
	cmd := &cobra.Command{
		Use:   "view <ID>",
		Short: "View a Smart Mirror server (Server/DC)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.ID = args[0]
			if runF != nil {
				return runF(opts)
			}
			return mirrorViewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Server/DC only)")
	return cmd
}

func mirrorViewRun(f *factory.Factory, opts *ViewOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	mc, err := backend.AsMirrorClient(client, host)
	if err != nil {
		return err
	}
	m, err := mc.GetMirrorServer(opts.ID)
	if err != nil {
		return err
	}

	if opts.Output.Format != "" {
		p := mirrorServerFields(f, opts.Output)
		p.AddItem(m)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "ID:      %s\n", m.ID)
	fmt.Fprintf(out, "Name:    %s\n", m.Name)
	fmt.Fprintf(out, "BaseURL: %s\n", m.BaseURL)
	fmt.Fprintf(out, "Enabled: %v\n", m.Enabled)
	return nil
}
