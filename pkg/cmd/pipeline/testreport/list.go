package testreport

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions holds parsed flags for `pipeline test-case list`.
type ListOptions struct {
	Output       format.OutputConfig
	Hostname     string
	PipelineUUID string
	StepUUID     string
	Status       string
	Limit        int
	Args         []string
}

// NewCmdList builds the `pipeline test-case list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list PIPELINE_UUID [PROJECT/REPO]",
		Short: "List test cases for a pipeline step",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			opts.PipelineUUID = args[0]
			opts.Args = args[1:]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.StepUUID, "step", "", "Step UUID (required)")
	_ = cmd.MarkFlagRequired("step")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status: PASSED, FAILED, or SKIPPED")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of test cases")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
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
	filter := backend.TestCaseFilter{
		Status: opts.Status,
		Limit:  opts.Limit,
	}
	cases, listErr := tc.ListPipelineTestCases(ref.Project, ref.Slug, opts.PipelineUUID, opts.StepUUID, filter)
	if listErr != nil && len(cases) == 0 {
		return listErr
	}

	p := format.New[backend.PipelineTestCase](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
	p.AddField(format.Field[backend.PipelineTestCase]{Name: "name", Header: "NAME", Extract: func(c backend.PipelineTestCase) any { return c.Name }})
	p.AddField(format.Field[backend.PipelineTestCase]{Name: "class_name", Header: "CLASS", Extract: func(c backend.PipelineTestCase) any { return c.ClassName }})
	p.AddField(format.Field[backend.PipelineTestCase]{Name: "status", Header: "STATUS", Extract: func(c backend.PipelineTestCase) any { return c.Status }})
	p.AddField(format.Field[backend.PipelineTestCase]{Name: "duration_ms", Header: "DURATION_MS", Extract: func(c backend.PipelineTestCase) any { return c.DurationMS }})
	p.AddField(format.Field[backend.PipelineTestCase]{Name: "failure_message", Header: "FAILURE", JSONOnly: true, Extract: func(c backend.PipelineTestCase) any { return c.FailureMessage }})
	for _, tc := range cases {
		p.AddItem(tc)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(cases), listErr)
	return listErr
}
