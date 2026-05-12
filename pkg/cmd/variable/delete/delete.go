// Package delete implements the `variable delete` subcommand.
package delete

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `variable delete`.
type Options struct {
	Hostname string
	Confirm  bool
	Scope    string // "repository" (default) | "workspace" | "deployment"
	EnvUUID  string // required when scope=deployment

	// Args[0] = PROJECT/REPO, Args[1] = KEY
	Args []string
}

// NewCmdDelete builds the `variable delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO KEY",
		Short: "Delete a variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Scope, "scope", "repository", "Variable scope: repository, workspace, or deployment")
	cmd.Flags().StringVar(&opts.EnvUUID, "env", "", "Environment UUID (required for --scope deployment)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	key := opts.Args[1]

	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required when not running interactively")
		}
		fmt.Fprintf(f.IOStreams.Out, "Delete variable %s? [y/N]: ", key)
		reader := bufio.NewReader(f.IOStreams.In)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "y" && answer != "Y" {
			fmt.Fprintln(f.IOStreams.Out, "Aborted.")
			return nil
		}
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
		if err := pc.DeletePipelineVariable(ref.Project, ref.Slug, key); err != nil {
			return err
		}

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, ref.Host)
		if err != nil {
			return err
		}
		if err := wc.DeleteWorkspaceVariable(ref.Project, key); err != nil {
			return err
		}

	case "deployment":
		if opts.EnvUUID == "" {
			return fmt.Errorf("--env ENV-UUID is required for --scope deployment")
		}
		dc, err := backend.AsDeploymentClient(client, ref.Host)
		if err != nil {
			return err
		}
		// Find by key first, then delete by UUID.
		vars, err := dc.ListEnvVariables(ref.Project, ref.Slug, opts.EnvUUID)
		if err != nil {
			return err
		}
		var varUUID string
		for _, v := range vars {
			if v.Key == key {
				varUUID = v.UUID
				break
			}
		}
		if varUUID == "" {
			return fmt.Errorf("variable %q not found in environment %s", key, opts.EnvUUID)
		}
		if err := dc.DeleteEnvVariable(ref.Project, ref.Slug, opts.EnvUUID, varUUID); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown scope %q; valid: repository, workspace, deployment", opts.Scope)
	}

	fmt.Fprintf(f.IOStreams.Out, "Deleted variable %s\n", key)
	return nil
}
