package perms

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

var validProjectPermissions = []string{"read", "write", "admin", "create-repo"}

// GrantOptions carries parsed flags for `workspace project perms grant`.
type GrantOptions struct {
	Hostname   string
	Workspace  string
	ProjectKey string
	User       string
	Group      string
	Permission string
}

// NewCmdGrant constructs the `workspace project perms grant` cobra command.
func NewCmdGrant(f *factory.Factory, runF func(*GrantOptions) error) *cobra.Command {
	opts := &GrantOptions{}
	cmd := &cobra.Command{
		Use:   "grant WORKSPACE PROJECT_KEY",
		Short: "Grant a user or group permission on a project (Cloud only)",
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
			return grantRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.User, "user", "", "User slug to grant permission to")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Group slug to grant permission to")
	cmd.Flags().StringVar(&opts.Permission, "permission", "", "Permission level: read, write, admin, create-repo (required)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("permission")
	return cmd
}

func grantRun(f *factory.Factory, opts *GrantOptions) error {
	if !isValidProjectPermission(opts.Permission) {
		return fmt.Errorf("invalid permission %q: must be one of read, write, admin, create-repo", opts.Permission)
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
	in := backend.WorkspaceProjectPermInput{Permission: opts.Permission}
	subject := opts.User
	if opts.Group != "" {
		in.GroupSlug = opts.Group
		subject = opts.Group
	} else {
		in.UserSlug = opts.User
	}
	if err := wpc.GrantWorkspaceProjectPerm(opts.Workspace, opts.ProjectKey, in); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Granted %s permission to %s on project %s in %s.\n",
		opts.Permission, subject, opts.ProjectKey, opts.Workspace)
	return nil
}

func isValidProjectPermission(p string) bool {
	for _, v := range validProjectPermissions {
		if p == v {
			return true
		}
	}
	return false
}
