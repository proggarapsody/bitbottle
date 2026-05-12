package steps

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/shared"
)

// Options holds parsed flags for `pipeline steps`.
type Options struct {
	Output   format.OutputConfig
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = PIPELINE-UUID
	Args []string
}

// NewCmdSteps builds the `pipeline steps` cobra command.
func NewCmdSteps(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "steps PROJECT/REPO PIPELINE-UUID",
		Short: "List the steps in a pipeline run",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return stepsRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func stepsRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
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
	steps, err := pc.ListPipelineSteps(ref.Project, ref.Slug, opts.Args[1])
	if err != nil {
		return err
	}
	p := shared.StepFields(f, opts.Output)
	for _, s := range steps {
		p.AddItem(s)
	}
	return p.Render()
}
