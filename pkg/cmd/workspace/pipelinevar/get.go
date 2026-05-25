package pipelinevar

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// GetOptions carries parsed flags for `workspace pipeline-variable get`.
type GetOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	Key       string
}

// NewCmdGet constructs the cobra command for `workspace pipeline-variable get`.
func NewCmdGet(f *factory.Factory, runF func(*GetOptions) error) *cobra.Command {
	opts := &GetOptions{}
	cmd := &cobra.Command{
		Use:   "get [WORKSPACE] KEY",
		Short: "Get a workspace pipeline variable by key",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(args) == 2 {
				opts.Workspace = args[0]
				opts.Key = args[1]
			} else {
				opts.Key = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return getRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func getRun(f *factory.Factory, opts *GetOptions) error {
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

	v, err := wpc.GetWorkspacePipelineVariable(workspace, uuid)
	if err != nil {
		return err
	}

	p := pipelineVarFields(f, opts.Output)
	p.AddItem(v)
	return p.Render()
}
