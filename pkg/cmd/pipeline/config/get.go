package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// GetOptions holds parsed flags for `pipeline config get`.
type GetOptions struct {
	Output   format.OutputConfig
	Hostname string
	Args     []string
}

// NewCmdGet builds the `pipeline config get` cobra command.
func NewCmdGet(f *factory.Factory, runF func(*GetOptions) error) *cobra.Command {
	opts := &GetOptions{}
	cmd := &cobra.Command{
		Use:   "get [PROJECT/REPO]",
		Short: "Get pipeline configuration for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runGet(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runGet(f *factory.Factory, _ *cobra.Command, opts *GetOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineConfigClient(client, ref.Host)
	if err != nil {
		return err
	}
	cfg, err := pc.GetPipelinesConfig(ref.Project, ref.Slug)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineConfig](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineConfig]{Name: "enabled", Header: "ENABLED", Extract: func(c backend.PipelineConfig) any { return c.Enabled }})
		p.SetSingleItem()
		p.AddItem(cfg)
		return p.Render()
	}

	fmt.Fprintf(f.IOStreams.Out, "Pipelines enabled: %v\n", cfg.Enabled)
	return nil
}
