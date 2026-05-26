package oidc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ConfigOptions holds parsed flags for `pipeline oidc config`.
type ConfigOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
}

// NewCmdConfig builds the `pipeline oidc config` cobra command.
func NewCmdConfig(f *factory.Factory, runF func(*ConfigOptions) error) *cobra.Command {
	opts := &ConfigOptions{}
	cmd := &cobra.Command{
		Use:   "config <WORKSPACE>",
		Short: "View the pipeline OIDC discovery document (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runConfig(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runConfig(f *factory.Factory, opts *ConfigOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	oc, err := backend.AsPipelineOIDCClient(client, host)
	if err != nil {
		return err
	}
	cfg, err := oc.GetPipelineOIDCConfig(opts.Workspace)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineOIDCConfig](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineOIDCConfig]{Name: "issuer", Header: "ISSUER", Extract: func(c backend.PipelineOIDCConfig) any { return c.Issuer }})
		p.AddField(format.Field[backend.PipelineOIDCConfig]{Name: "jwks_uri", Header: "JWKS_URI", Extract: func(c backend.PipelineOIDCConfig) any { return c.JWKSURI }})
		p.AddField(format.Field[backend.PipelineOIDCConfig]{Name: "claims_supported", Header: "CLAIMS_SUPPORTED", Extract: func(c backend.PipelineOIDCConfig) any { return strings.Join(c.ClaimsSupported, ",") }})
		p.SetSingleItem()
		p.AddItem(cfg)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "issuer=%s\n", cfg.Issuer)
	fmt.Fprintf(out, "jwks_uri=%s\n", cfg.JWKSURI)
	if len(cfg.ClaimsSupported) > 0 {
		fmt.Fprintf(out, "claims_supported=%s\n", strings.Join(cfg.ClaimsSupported, ","))
	}
	return nil
}
