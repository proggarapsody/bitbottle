// Package stop implements `bitbottle pipeline stop UUID [PROJECT/REPO]`.
package stop

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline stop`.
type Options struct {
	Hostname string
	Confirm  bool

	// Args[0] = UUID; Args[1] = optional PROJECT/REPO
	Args []string
}

// NewCmdStop builds the `pipeline stop` cobra command.
func NewCmdStop(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "stop UUID [PROJECT/REPO]",
		Short: "Stop a running pipeline",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runStop(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm stopping the pipeline without a TTY prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runStop(f *factory.Factory, opts *Options) error {
	// Non-TTY guard: require --confirm when not running interactively.
	if !f.IOStreams.IsStdoutTTY() && !opts.Confirm {
		return fmt.Errorf("pass --confirm to stop a running pipeline")
	}

	// UUID is always args[0]; optional PROJECT/REPO is args[1].
	uuid := opts.Args[0]
	repoArgs := opts.Args[1:]

	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineClient(client, ref.Host)
	if err != nil {
		return err
	}

	if err := pc.StopPipeline(ref.Project, ref.Slug, uuid); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Stopped pipeline %s\n", uuid)
	return nil
}
