package cache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions holds parsed flags for `pipeline cache delete`.
type DeleteOptions struct {
	Hostname string
	UUID     string
	Args     []string
}

// NewCmdDelete builds the `pipeline cache delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] UUID",
		Short: "Delete a pipeline cache",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// args is either [UUID] or [PROJECT/REPO, UUID]
			if len(args) == 1 {
				opts.UUID = args[0]
				opts.Args = nil
			} else {
				opts.Args = args[:1]
				opts.UUID = args[1]
			}
			if runF != nil {
				return runF(opts)
			}
			return runDelete(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runDelete(f *factory.Factory, opts *DeleteOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineCacheClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := pc.DeletePipelineCache(ref.Project, ref.Slug, opts.UUID); err != nil {
		return err
	}
	out := f.IOStreams.Out
	fmt.Fprintf(out, "Deleted pipeline cache %s\n", opts.UUID)
	return nil
}
