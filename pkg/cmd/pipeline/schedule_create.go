package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ScheduleCreateOptions holds parsed flags for `pipeline schedule create`.
type ScheduleCreateOptions struct {
	Hostname string
	Cron     string
	Branch   string
	Enabled  bool
	Args     []string
}

// NewCmdScheduleCreate builds the `pipeline schedule create` cobra command.
func NewCmdScheduleCreate(f *factory.Factory, runF func(*ScheduleCreateOptions) error) *cobra.Command {
	opts := &ScheduleCreateOptions{
		Enabled: true,
	}
	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO]",
		Short: "Create a pipeline schedule for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runScheduleCreate(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&opts.Cron, "cron", "", "Cron expression for the schedule (e.g. \"0 0 * * *\")")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Branch to run the pipeline on")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", true, "Whether the schedule is enabled")
	_ = cmd.MarkFlagRequired("cron")
	_ = cmd.MarkFlagRequired("branch")
	return cmd
}

func runScheduleCreate(f *factory.Factory, cmd *cobra.Command, opts *ScheduleCreateOptions) error {
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
	created, err := sc.CreatePipelineSchedule(ref.Project, ref.Slug, backend.PipelineScheduleInput{
		CronExpression: opts.Cron,
		Branch:         opts.Branch,
		Enabled:        opts.Enabled,
	})
	if err != nil {
		return err
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.PipelineSchedule](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.SetSingleItem()
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "uuid", Header: "UUID", Extract: func(s backend.PipelineSchedule) any { return s.UUID }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "cronExpression", Header: "CRON", Extract: func(s backend.PipelineSchedule) any { return s.CronExpression }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "branch", Header: "BRANCH", Extract: func(s backend.PipelineSchedule) any { return s.Branch }})
		p.AddField(format.Field[backend.PipelineSchedule]{Name: "enabled", Header: "ENABLED", Extract: func(s backend.PipelineSchedule) any { return s.Enabled }})
		p.AddItem(created)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Schedule created (UUID: %s, cron: %s, branch: %s, enabled: %v)\n",
		created.UUID, created.CronExpression, created.Branch, created.Enabled)
	return nil
}
