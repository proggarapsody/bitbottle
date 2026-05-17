package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ScheduleListOptions holds parsed flags for `pipeline schedule list`.
type ScheduleListOptions struct {
	Hostname string
	Args     []string
}

// NewCmdScheduleList builds the `pipeline schedule list` cobra command.
func NewCmdScheduleList(f *factory.Factory, runF func(*ScheduleListOptions) error) *cobra.Command {
	opts := &ScheduleListOptions{}
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pipeline schedules for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runScheduleList(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runScheduleList(f *factory.Factory, cmd *cobra.Command, opts *ScheduleListOptions) error {
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
	schedules, listErr := sc.ListPipelineSchedules(ref.Project, ref.Slug)
	if listErr != nil && len(schedules) == 0 {
		return listErr
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.PipelineSchedule](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "uuid", Header: "UUID", Extract: func(s backend.PipelineSchedule) any { return s.UUID }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "cronExpression", Header: "CRON", Extract: func(s backend.PipelineSchedule) any { return s.CronExpression }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "branch", Header: "BRANCH", Extract: func(s backend.PipelineSchedule) any { return s.Branch }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "enabled", Header: "ENABLED", Extract: func(s backend.PipelineSchedule) any { return s.Enabled }})
		for _, s := range schedules {
			p.AddItem(s)
		}
		if err := p.Render(); err != nil {
			return err
		}
		cmdutil.PartialWarn(f.IOStreams.ErrOut, len(schedules), listErr)
		return listErr
	}

	out := f.IOStreams.Out
	if len(schedules) == 0 {
		fmt.Fprintln(out, "No pipeline schedules found.")
		return nil
	}
	fmt.Fprintf(out, "%-38s  %-20s  %-20s  %s\n", "UUID", "CRON", "BRANCH", "ENABLED")
	for _, s := range schedules {
		fmt.Fprintf(out, "%-38s  %-20s  %-20s  %v\n", s.UUID, s.CronExpression, s.Branch, s.Enabled)
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(schedules), listErr)
	return listErr
}
