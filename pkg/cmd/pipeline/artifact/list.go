package artifact

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions holds parsed flags for `pipeline artifact list`.
type ListOptions struct {
	Output       format.OutputConfig
	Hostname     string
	Limit        int
	PipelineUUID string
	StepUUID     string
	Args         []string
}

// NewCmdList builds the `pipeline artifact list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list PIPELINE_UUID [PROJECT/REPO]",
		Short: "List artifacts for a pipeline step",
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
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of artifacts")
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
	ac, err := backend.AsPipelineArtifactClient(client, ref.Host)
	if err != nil {
		return err
	}
	artifacts, listErr := ac.ListPipelineArtifacts(ref.Project, ref.Slug, opts.PipelineUUID, opts.StepUUID, opts.Limit)
	if listErr != nil && len(artifacts) == 0 {
		return listErr
	}
	p := format.New[backend.PipelineArtifact](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
	p.AddField(format.Field[backend.PipelineArtifact]{Name: "name", Header: "NAME", Extract: func(a backend.PipelineArtifact) any { return a.Name }})
	p.AddField(format.Field[backend.PipelineArtifact]{Name: "size_bytes", Header: "SIZE", Extract: func(a backend.PipelineArtifact) any { return a.SizeBytes }})
	p.AddField(format.Field[backend.PipelineArtifact]{Name: "url", Header: "URL", JSONOnly: true, Extract: func(a backend.PipelineArtifact) any { return a.URL }})
	for _, a := range artifacts {
		p.AddItem(a)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(artifacts), listErr)
	return listErr
}
