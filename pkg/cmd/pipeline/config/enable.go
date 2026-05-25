package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// EnableOptions holds parsed flags for `pipeline config enable`.
type EnableOptions struct {
	Hostname string
	Args     []string
}

// NewCmdEnable builds the `pipeline config enable` cobra command.
func NewCmdEnable(f *factory.Factory, runF func(*EnableOptions) error) *cobra.Command {
	opts := &EnableOptions{}
	cmd := &cobra.Command{
		Use:   "enable [PROJECT/REPO]",
		Short: "Enable pipelines for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runEnable(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runEnable(f *factory.Factory, opts *EnableOptions) error {
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
	if _, err := pc.UpdatePipelinesConfig(ref.Project, ref.Slug, backend.PipelineConfig{Enabled: true}); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Pipelines enabled for %s/%s.\n", ref.Project, ref.Slug)
	return nil
}
