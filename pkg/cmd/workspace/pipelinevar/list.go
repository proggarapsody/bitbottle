package pipelinevar

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `workspace pipeline-variable list`.
type ListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
}

// NewCmdList constructs the cobra command for `workspace pipeline-variable list`.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List workspace pipeline variables",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(args) > 0 {
				opts.Workspace = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
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
	vars, listErr := wpc.ListWorkspacePipelineVariables(workspace)
	if listErr != nil && len(vars) == 0 {
		return listErr
	}

	p := pipelineVarFields(f, opts.Output)
	for _, v := range vars {
		p.AddItem(v)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(vars), listErr)
	return listErr
}

func pipelineVarFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PipelineVariable] {
	p := format.New[backend.PipelineVariable](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.PipelineVariable]{
		Name:    "uuid",
		Header:  "UUID",
		JSONOnly: true,
		Extract: func(v backend.PipelineVariable) any { return v.UUID },
	})
	p.AddField(format.Field[backend.PipelineVariable]{
		Name:    "key",
		Header:  "KEY",
		Extract: func(v backend.PipelineVariable) any { return v.Key },
	})
	p.AddField(format.Field[backend.PipelineVariable]{
		Name:    "value",
		Header:  "VALUE",
		Extract: func(v backend.PipelineVariable) any {
			if v.Secured {
				return "***"
			}
			return v.Value
		},
	})
	p.AddField(format.Field[backend.PipelineVariable]{
		Name:    "secured",
		Header:  "SECURED",
		Extract: func(v backend.PipelineVariable) any { return v.Secured },
	})
	return p
}
