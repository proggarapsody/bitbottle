package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/shared"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment variable list`.
type Options struct {
	Hostname   string
	JSONFields string
	JQExpr     string

	// Args[0] = PROJECT/REPO, Args[1] = ENV-UUID
	Args []string
}

// NewCmdList builds the `environment variable list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO ENV-UUID",
		Short: "List environment variables",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
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
	vars, err := dc.ListEnvVariables(ref.Project, ref.Slug, opts.Args[1])
	if err != nil {
		return err
	}
	p := shared.EnvVariableFields(f, opts.JSONFields, opts.JQExpr)
	for _, v := range vars {
		p.AddItem(v)
	}
	return p.Render()
}
