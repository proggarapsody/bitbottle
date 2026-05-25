package mirror

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `mirror list`.
type ListOptions struct {
	Output   format.OutputConfig
	Hostname string
	Limit    int
}

// NewCmdMirrorList constructs the `mirror list` cobra command.
func NewCmdMirrorList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Smart Mirror servers (Server/DC)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return mirrorListRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of mirrors (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Server/DC only)")
	return cmd
}

func mirrorListRun(f *factory.Factory, opts *ListOptions) error {
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
	mirrors, listErr := mc.ListMirrorServers(opts.Limit)
	if listErr != nil && len(mirrors) == 0 {
		return listErr
	}

	p := mirrorServerFields(f, opts.Output)
	for _, m := range mirrors {
		p.AddItem(m)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(mirrors), listErr)
	return listErr
}

func mirrorServerFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.MirrorServer] {
	p := format.New[backend.MirrorServer](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.MirrorServer]{
		Name:    "id",
		Header:  "ID",
		Extract: func(m backend.MirrorServer) any { return m.ID },
	})
	p.AddField(format.Field[backend.MirrorServer]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(m backend.MirrorServer) any { return m.Name },
	})
	p.AddField(format.Field[backend.MirrorServer]{
		Name:    "base_url",
		Header:  "BASE URL",
		Extract: func(m backend.MirrorServer) any { return m.BaseURL },
	})
	p.AddField(format.Field[backend.MirrorServer]{
		Name:    "enabled",
		Header:  "ENABLED",
		Extract: func(m backend.MirrorServer) any { return m.Enabled },
	})
	return p
}
