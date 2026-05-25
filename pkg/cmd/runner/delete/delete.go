// Package delete implements the `runner delete` command.
package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions holds parsed flags for `runner delete`.
type DeleteOptions struct {
	Hostname  string
	Workspace string
	UUID      string
}

// NewCmdDelete builds the `runner delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [WORKSPACE] RUNNER_UUID",
		Short: "Remove a self-hosted runner",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 2 {
				opts.Workspace = args[0]
				opts.UUID = args[1]
			} else {
				opts.UUID = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return runDelete(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func runDelete(f *factory.Factory, cmd *cobra.Command, opts *DeleteOptions) error {
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

	rc, err := backend.AsRunnerClient(client, host)
	if err != nil {
		return err
	}

	if err := rc.DeleteRunner(workspace, opts.UUID); err != nil {
		return err
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.Runner](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.AddField(format.Field[backend.Runner]{Name: "uuid", Header: "UUID", Extract: func(r backend.Runner) any { return r.UUID }})
		p.AddItem(backend.Runner{UUID: opts.UUID})
		return p.Render()
	}

	fmt.Fprintf(f.IOStreams.Out, "Deleted runner %s\n", opts.UUID)
	return nil
}

// resolveWorkspace returns the workspace slug from the explicit arg, or falls
// back to the pinned repo's namespace. An error is returned when neither is available.
func resolveWorkspace(f *factory.Factory, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	ref, err := f.BaseRepo()
	if err == nil && ref.Project != "" {
		return ref.Project, nil
	}
	return "", fmt.Errorf("workspace required: pass a workspace slug as an argument or run from inside a Cloud checkout")
}
