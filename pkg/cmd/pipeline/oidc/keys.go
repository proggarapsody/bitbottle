package oidc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// KeysOptions holds parsed flags for `pipeline oidc keys`.
type KeysOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
}

// NewCmdKeys builds the `pipeline oidc keys` cobra command.
func NewCmdKeys(f *factory.Factory, runF func(*KeysOptions) error) *cobra.Command {
	opts := &KeysOptions{}
	cmd := &cobra.Command{
		Use:   "keys <WORKSPACE>",
		Short: "View the pipeline OIDC JWKS key set (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runKeys(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runKeys(f *factory.Factory, opts *KeysOptions) error {
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
	keys, err := oc.GetPipelineOIDCKeys(opts.Workspace)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineOIDCKeys](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineOIDCKeys]{Name: "keys", Header: "KEYS", Extract: func(k backend.PipelineOIDCKeys) any { return k.Keys }})
		p.SetSingleItem()
		p.AddItem(keys)
		return p.Render()
	}

	out := f.IOStreams.Out
	for _, k := range keys.Keys {
		fmt.Fprintf(out, "kid=%s  kty=%s  alg=%s\n", k.Kid, k.Kty, k.Alg)
	}
	return nil
}
