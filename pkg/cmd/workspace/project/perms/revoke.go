package perms

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// RevokeOptions carries parsed flags for `workspace project perms revoke`.
type RevokeOptions struct {
	Hostname   string
	Workspace  string
	ProjectKey string
	User       string
	Group      string
	Confirm    bool
}

// NewCmdRevoke constructs the `workspace project perms revoke` cobra command.
func NewCmdRevoke(f *factory.Factory, runF func(*RevokeOptions) error) *cobra.Command {
	opts := &RevokeOptions{}
	cmd := &cobra.Command{
		Use:   "revoke WORKSPACE PROJECT_KEY",
		Short: "Revoke a user or group permission on a project (Cloud only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			opts.ProjectKey = args[1]
			if opts.User == "" && opts.Group == "" {
				return fmt.Errorf("must specify --user or --group")
			}
			if opts.User != "" && opts.Group != "" {
				return fmt.Errorf("--user and --group are mutually exclusive")
			}
			if runF != nil {
				return runF(opts)
			}
			return revokeRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.User, "user", "", "User slug to revoke permission from")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Group slug to revoke permission from")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func revokeRun(f *factory.Factory, opts *RevokeOptions) error {
	if !opts.Confirm && !f.IOStreams.IsStdoutTTY() {
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
	wpc, err := backend.AsWorkspaceProjectPermsClient(client, host)
	if err != nil {
		return err
	}
	isGroup := opts.Group != ""
	slug := opts.User
	if isGroup {
		slug = opts.Group
	}
	if err := wpc.RevokeWorkspaceProjectPerm(opts.Workspace, opts.ProjectKey, slug, isGroup); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Revoked %s's permission on project %s in %s.\n",
		slug, opts.ProjectKey, opts.Workspace)
	return nil
}
