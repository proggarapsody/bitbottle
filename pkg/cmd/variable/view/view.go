// Package view implements the `variable view` subcommand.
package view

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

// Options holds parsed flags for `variable view`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Scope    string // "repository" (default) | "workspace" | "deployment"
	EnvUUID  string // required when scope=deployment

	// Args[0] = PROJECT/REPO (optional), Args[last] = KEY
	Args []string
}

// NewCmdView builds the `variable view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO] KEY",
		Short: "View a pipeline variable by key",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Scope, "scope", "repository", "Variable scope: repository, workspace, or deployment")
	cmd.Flags().StringVar(&opts.EnvUUID, "env", "", "Environment UUID (required for --scope deployment)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 1)
	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	key := rest[0]

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}

	ops, err := shared.ResolveVariableOps(opts.Scope, client, ref.Host, ref.Project, ref.Slug, opts.EnvUUID)
	if err != nil {
		return err
	}
	item, err := ops.GetVariableByKey(key)
	if err != nil {
		return err
	}

	if opts.Scope == "deployment" {
		p := shared.EnvVariableFields(f, opts.Output)
		p.AddItem(backend.EnvVariable{UUID: item.UUID, Key: item.Key, Value: item.Value, Secured: item.Secured})
		return p.Render()
	}
	p := shared.VariableFields(f, opts.Output)
	p.AddItem(backend.PipelineVariable{UUID: item.UUID, Key: item.Key, Value: item.Value, Secured: item.Secured})
	return p.Render()
}
