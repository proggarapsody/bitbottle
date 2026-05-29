package info

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdHostInfo(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show backend type, version, and supported features",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			hic, err := backend.AsHostInfoClient(client, host)
			if err != nil {
				return err
			}

			info, err := hic.GetHostInfo()
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := hostInfoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(info)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Backend:  %s\n", info.BackendType)
			fmt.Fprintf(f.IOStreams.Out, "Host:     %s\n", host)
			if info.Version != "" {
				fmt.Fprintf(f.IOStreams.Out, "Version:  %s\n", info.Version)
			}
			if len(info.SupportedFeatures) > 0 {
				fmt.Fprintf(f.IOStreams.Out, "Features: %s\n", strings.Join(info.SupportedFeatures, ", "))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func hostInfoFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.HostInfo] {
	p := format.New[backend.HostInfo](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "backend_type",
		Header:  "BACKEND",
		Extract: func(h backend.HostInfo) any { return h.BackendType },
	})
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "base_url",
		Header:  "URL",
		Extract: func(h backend.HostInfo) any { return h.BaseURL },
	})
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "version",
		Header:  "VERSION",
		Extract: func(h backend.HostInfo) any { return h.Version },
	})
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "build_number",
		Header:  "BUILD",
		Extract: func(h backend.HostInfo) any { return h.BuildNumber },
	})
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "display_name",
		Header:  "NAME",
		Extract: func(h backend.HostInfo) any { return h.DisplayName },
	})
	p.AddField(format.Field[backend.HostInfo]{
		Name:    "supported_features",
		Header:  "FEATURES",
		Extract: func(h backend.HostInfo) any { return h.SupportedFeatures },
	})
	return p
}
