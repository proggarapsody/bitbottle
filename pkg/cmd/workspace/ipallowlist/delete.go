package ipallowlist

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions carries parsed flags for `workspace ipallowlist delete`.
type DeleteOptions struct {
	Hostname  string
	Workspace string
	UUID      string
	Confirm   bool
}

// NewCmdDelete constructs the cobra command for `workspace ipallowlist delete`.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete [WORKSPACE] UUID",
		Short: "Delete an IP allowlist entry from a workspace (Cloud only)",
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
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	return cmd
}

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
	workspace, err := resolveWorkspace(f, opts.Workspace)
	if err != nil {
		return err
	}

	if !f.IOStreams.IsStdoutTTY() && !opts.Confirm {
		return fmt.Errorf("--confirm required when not running interactively")
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	ac, err := backend.AsIPAllowlistClient(client, host)
	if err != nil {
		return err
	}

	if err := ac.DeleteIPAllowlist(workspace, opts.UUID); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Deleted IP allowlist entry %s\n", opts.UUID)
	return nil
}
