// Package list implements the `variable list` subcommand.
package list

import (
	"fmt"

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

	switch opts.Scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, ref.Host)
		if err != nil {
			return err
		}
		vars, err := pc.ListPipelineVariables(ref.Project, ref.Slug)
		if err != nil {
			return err
		}
		p := shared.VariableFields(f, opts.Output)
		for _, v := range vars {
			p.AddItem(v)
		}
		return p.Render()

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, ref.Host)
		if err != nil {
			return err
		}
		vars, err := wc.ListWorkspaceVariables(ref.Project)
		if err != nil {
			return err
		}
		p := shared.VariableFields(f, opts.Output)
		for _, v := range vars {
			p.AddItem(v)
		}
		return p.Render()

	case "deployment":
		if opts.EnvUUID == "" {
			return fmt.Errorf("--env ENV-UUID is required for --scope deployment")
		}
		dc, err := backend.AsDeploymentClient(client, ref.Host)
		if err != nil {
			return err
		}
		vars, err := dc.ListEnvVariables(ref.Project, ref.Slug, opts.EnvUUID)
		if err != nil {
			return err
		}
		p := shared.EnvVariableFields(f, opts.Output)
		for _, v := range vars {
			p.AddItem(v)
		}
		return p.Render()

	default:
		return fmt.Errorf("unknown scope %q; valid: repository, workspace, deployment", opts.Scope)
	}
}
