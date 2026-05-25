package pipelinevar

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions carries parsed flags for `workspace pipeline-variable delete`.
type DeleteOptions struct {
	Hostname  string
	Workspace string
	Key       string
	Confirm   bool
}

// NewCmdDelete constructs the cobra command for `workspace pipeline-variable delete`.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [WORKSPACE] KEY",
		Short: "Delete a workspace pipeline variable by key",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 {
				opts.Workspace = args[0]
				opts.Key = args[1]
			} else {
				opts.Key = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	return cmd
}

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
	if !opts.Confirm && !f.IOStreams.IsStdoutTTY() {
		return fmt.Errorf("--confirm required when not running interactively")
	}

	workspace, err := resolveWorkspace(f, opts.Workspace)
	if err != nil {
		return err
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wpc, err := backend.AsWorkspacePipelineVariableClient(client, host)
	if err != nil {
		return err
	}

	// Resolve key → UUID via list.
	vars, err := wpc.ListWorkspacePipelineVariables(workspace)
	if err != nil {
		return err
	}
	var uuid string
	for _, v := range vars {
		if v.Key == opts.Key {
			uuid = v.UUID
			break
		}
	}
	if uuid == "" {
		return &backend.DomainError{
			Kind:     backend.ErrNotFound,
			Resource: "workspace-pipeline-variable",
			ID:       opts.Key,
			Message:  fmt.Sprintf("workspace pipeline variable %q not found", opts.Key),
		}
	}

	if err := wpc.DeleteWorkspacePipelineVariable(workspace, uuid); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Deleted workspace pipeline variable %q.\n", opts.Key)
	return nil
}
