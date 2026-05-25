package perms

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// RevokeOptions carries parsed flags for `workspace perms revoke`.
type RevokeOptions struct {
	Hostname  string
	Workspace string
	User      string
	Confirm   bool
}

// NewCmdWorkspacePermsRevoke constructs the `workspace perms revoke` command.
func NewCmdWorkspacePermsRevoke(f *factory.Factory, runF func(*RevokeOptions) error) *cobra.Command {
	opts := &RevokeOptions{}
	cmd := &cobra.Command{
		Use:   "revoke <WORKSPACE>",
		Short: "Revoke a user's permission in a workspace (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			if runF != nil {
				return runF(opts)
			}
			return workspacePermsRevokeRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.User, "user", "", "User to revoke permission from (required)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

func workspacePermsRevokeRun(f *factory.Factory, opts *RevokeOptions) error {
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
	wpc, err := backend.AsWorkspacePermsClient(client, host)
	if err != nil {
		return err
	}
	if err := wpc.RevokeWorkspacePerm(opts.Workspace, opts.User); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Revoked %s's permission in %s.\n", opts.User, opts.Workspace)
	return nil
}
