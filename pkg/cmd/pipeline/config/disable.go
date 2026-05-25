package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DisableOptions holds parsed flags for `pipeline config disable`.
type DisableOptions struct {
	Hostname string
	Args     []string
}

// NewCmdDisable builds the `pipeline config disable` cobra command.
func NewCmdDisable(f *factory.Factory, runF func(*DisableOptions) error) *cobra.Command {
	opts := &DisableOptions{}
	cmd := &cobra.Command{
		Use:   "disable [PROJECT/REPO]",
		Short: "Disable pipelines for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runDisable(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runDisable(f *factory.Factory, opts *DisableOptions) error {
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
	if _, err := pc.UpdatePipelinesConfig(ref.Project, ref.Slug, backend.PipelineConfig{Enabled: false}); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Pipelines disabled for %s/%s.\n", ref.Project, ref.Slug)
	return nil
}
