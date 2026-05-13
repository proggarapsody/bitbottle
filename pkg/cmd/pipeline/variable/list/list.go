package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	varshared "github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

// Options holds parsed flags for `pipeline variable list`.
type Options struct {
	Output   format.OutputConfig
	Hostname string

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `pipeline variable list` cobra command.
//
// Deprecated: use `bitbottle variable list --scope repository` instead.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:        "list PROJECT/REPO",
		Short:      "List pipeline variables",
		Deprecated: "use `bitbottle variable list --scope repository` instead",
		Args:       cobra.ExactArgs(1),
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
	ops, err := varshared.ResolveVariableOps("repository", client, ref.Host, ref.Project, ref.Slug, "")
	if err != nil {
		return err
	}
	vars, err := ops.ListVariables()
	if err != nil {
		return err
	}
	p := varshared.VariableFields(f, opts.Output)
	for _, v := range vars {
		p.AddItem(backend.PipelineVariable{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
	}
	return p.Render()
}
