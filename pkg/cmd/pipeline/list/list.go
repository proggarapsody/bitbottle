package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/shared"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options holds parsed flags for `pipeline list`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Limit    int

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `pipeline list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List pipelines",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Maximum number of pipelines")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
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
	pipelines, err := pc.ListPipelines(ref.Project, ref.Slug, opts.Limit)
	if err != nil {
		return err
	}
	p := shared.PipelineFields(f, opts.Output)
	for _, pl := range pipelines {
		p.AddItem(pl)
	}
	return p.Render()
}
