package testreport

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ViewOptions holds parsed flags for `pipeline test-report view`.
type ViewOptions struct {
	Output       format.OutputConfig
	Hostname     string
	PipelineUUID string
	StepUUID     string
	Args         []string
}

// NewCmdView builds the `pipeline test-report view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{}
	cmd := &cobra.Command{
		Use:   "view PIPELINE_UUID [PROJECT/REPO]",
		Short: "View test report summary for a pipeline step",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.PipelineUUID = args[0]
			opts.Args = args[1:]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.StepUUID, "step", "", "Step UUID (required)")
	_ = cmd.MarkFlagRequired("step")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func viewRun(f *factory.Factory, opts *ViewOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	tc, err := backend.AsPipelineTestReportClient(client, ref.Host)
	if err != nil {
		return err
	}
	report, err := tc.GetPipelineTestReport(ref.Project, ref.Slug, opts.PipelineUUID, opts.StepUUID)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineTestReport](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineTestReport]{Name: "total", Header: "TOTAL", Extract: func(r backend.PipelineTestReport) any { return r.Total }})
		p.AddField(format.Field[backend.PipelineTestReport]{Name: "passed", Header: "PASSED", Extract: func(r backend.PipelineTestReport) any { return r.Passed }})
		p.AddField(format.Field[backend.PipelineTestReport]{Name: "failed", Header: "FAILED", Extract: func(r backend.PipelineTestReport) any { return r.Failed }})
		p.AddField(format.Field[backend.PipelineTestReport]{Name: "skipped", Header: "SKIPPED", Extract: func(r backend.PipelineTestReport) any { return r.Skipped }})
		p.AddField(format.Field[backend.PipelineTestReport]{Name: "duration_ms", Header: "DURATION_MS", Extract: func(r backend.PipelineTestReport) any { return r.DurationMS }})
		p.SetSingleItem()
		p.AddItem(report)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Total:   %d\n", report.Total)
	fmt.Fprintf(out, "Passed:  %d\n", report.Passed)
	fmt.Fprintf(out, "Failed:  %d\n", report.Failed)
	fmt.Fprintf(out, "Skipped: %d\n", report.Skipped)
	fmt.Fprintf(out, "Duration: %dms\n", report.DurationMS)
	return nil
}
