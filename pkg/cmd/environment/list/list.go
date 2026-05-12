package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/shared"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment list`.
type Options struct {
	Output   format.OutputConfig
	Hostname string

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `environment list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List deployment environments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
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
	dc, err := backend.AsDeploymentClient(client, ref.Host)
	if err != nil {
		return err
	}
	envs, err := dc.ListEnvironments(ref.Project, ref.Slug)
	if err != nil {
		return err
	}
	p := shared.EnvironmentFields(f, opts.Output)
	for _, e := range envs {
		p.AddItem(e)
	}
	return p.Render()
}
