// Package list implements the `variable list` subcommand.
package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

// Options holds parsed flags for `variable list`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Scope    string // "repository" (default) | "workspace" | "deployment"
	EnvUUID  string // required when scope=deployment

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `variable list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List variables (repository, workspace, or deployment scope)",
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
	cmd.Flags().StringVar(&opts.Scope, "scope", "repository", "Variable scope: repository, workspace, or deployment")
	cmd.Flags().StringVar(&opts.EnvUUID, "env", "", "Environment UUID (required for --scope deployment)")
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

	ops, err := shared.ResolveVariableOps(opts.Scope, client, ref.Host, ref.Project, ref.Slug, opts.EnvUUID)
	if err != nil {
		return err
	}
	vars, err := ops.ListVariables()
	if err != nil {
		return err
	}

	// Pick the printer based on scope: deployment variables surface as
	// backend.EnvVariable for JSON/template parity with the old switch;
	// repository and workspace surface as backend.PipelineVariable.
	if opts.Scope == "deployment" {
		p := shared.EnvVariableFields(f, opts.Output)
		for _, v := range vars {
			p.AddItem(backend.EnvVariable{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
		}
		return p.Render()
	}
	p := shared.VariableFields(f, opts.Output)
	for _, v := range vars {
		p.AddItem(backend.PipelineVariable{UUID: v.UUID, Key: v.Key, Value: v.Value, Secured: v.Secured})
	}
	return p.Render()
}
