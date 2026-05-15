package hook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions carries parsed flags for `workspace hook delete`.
type DeleteOptions struct {
	Hostname  string
	Workspace string
	UUID      string
}

// NewCmdDelete constructs the cobra command for `workspace hook delete`.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [WORKSPACE] UUID",
		Short: "Delete a workspace-level webhook",
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
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
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
	wwc, err := backend.AsWorkspaceWebhookClient(client, host)
	if err != nil {
		return err
	}

	if err := wwc.DeleteWorkspaceWebhook(workspace, opts.UUID); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Deleted workspace webhook %s.\n", opts.UUID)
	return nil
}
