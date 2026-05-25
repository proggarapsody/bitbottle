package pipelinevar

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// SetOptions carries parsed flags for `workspace pipeline-variable set`.
type SetOptions struct {
	Hostname  string
	Workspace string
	Key       string
	Value     string
	Secured   bool
}

// NewCmdSet constructs the cobra command for `workspace pipeline-variable set`.
func NewCmdSet(f *factory.Factory, runF func(*SetOptions) error) *cobra.Command {
	opts := &SetOptions{}
	cmd := &cobra.Command{
		Use:   "set [WORKSPACE] KEY VALUE",
		Short: "Create or update a workspace pipeline variable",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 3 {
				opts.Workspace = args[0]
				opts.Key = args[1]
				opts.Value = args[2]
			} else {
				opts.Key = args[0]
				opts.Value = args[1]
			}
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().BoolVar(&opts.Secured, "secured", false, "Mark as secured (value redacted on read)")
	return cmd
}

func setRun(f *factory.Factory, opts *SetOptions) error {
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

	v, err := wpc.SetWorkspacePipelineVariable(workspace, backend.PipelineVariableInput{
		Key:     opts.Key,
		Value:   opts.Value,
		Secured: opts.Secured,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Set workspace pipeline variable %q (UUID: %s).\n", v.Key, v.UUID)
	return nil
}
