package perms

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

var validPermissions = []string{"member", "collaborator", "owner"}

// GrantOptions carries parsed flags for `workspace perms grant`.
type GrantOptions struct {
	Hostname   string
	Workspace  string
	User       string
	Permission string
}

// NewCmdWorkspacePermsGrant constructs the `workspace perms grant` command.
func NewCmdWorkspacePermsGrant(f *factory.Factory, runF func(*GrantOptions) error) *cobra.Command {
	opts := &GrantOptions{}
	cmd := &cobra.Command{
		Use:   "grant <WORKSPACE>",
		Short: "Grant a user permission in a workspace (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			if runF != nil {
				return runF(opts)
			}
			return workspacePermsGrantRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.User, "user", "", "User to grant permission to (required)")
	cmd.Flags().StringVar(&opts.Permission, "permission", "", "Permission level: member, collaborator, or owner (required)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("permission")
	return cmd
}

func workspacePermsGrantRun(f *factory.Factory, opts *GrantOptions) error {
	if !isValidPermission(opts.Permission) {
		return fmt.Errorf("invalid permission %q: must be one of member, collaborator, owner", opts.Permission)
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
	if err := wpc.GrantWorkspacePerm(opts.Workspace, opts.User, opts.Permission); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Granted %s permission to %s in %s.\n", opts.Permission, opts.User, opts.Workspace)
	return nil
}

func isValidPermission(p string) bool {
	for _, v := range validPermissions {
		if p == v {
			return true
		}
	}
	return false
}
