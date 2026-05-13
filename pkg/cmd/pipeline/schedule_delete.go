package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ScheduleDeleteOptions holds parsed flags for `pipeline schedule delete`.
type ScheduleDeleteOptions struct {
	Hostname string
	UUID     string
	Args     []string
}

// NewCmdScheduleDelete builds the `pipeline schedule delete` cobra command.
func NewCmdScheduleDelete(f *factory.Factory, runF func(*ScheduleDeleteOptions) error) *cobra.Command {
	opts := &ScheduleDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] UUID",
		Short: "Delete a pipeline schedule from a repository",
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
			return runScheduleDelete(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runScheduleDelete(f *factory.Factory, opts *ScheduleDeleteOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	sc, err := backend.AsPipelineScheduleClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := sc.DeletePipelineSchedule(ref.Project, ref.Slug, opts.UUID); err != nil {
		return err
	}
	out := f.IOStreams.Out
	fmt.Fprintf(out, "Schedule %s deleted\n", opts.UUID)
	return nil
}
