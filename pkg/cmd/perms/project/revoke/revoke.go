package revoke

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/shared"
)

// Options holds parsed flags for `perms project revoke`.
type Options struct {
	Hostname  string
	UserSlug  string
	GroupName string
	// Args[0] = PROJECT
	Args []string
}

// NewCmdRevoke builds the `perms project revoke` cobra command.
func NewCmdRevoke(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "revoke PROJECT",
		Short: "Revoke a permission on a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return revokeRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&opts.UserSlug, "user", "", "User slug to revoke the permission from")
	cmd.Flags().StringVar(&opts.GroupName, "group", "", "Group name to revoke the permission from")
	return cmd
}

func revokeRun(f *factory.Factory, opts *Options) error {
	subject, err := shared.SubjectFromFlags(opts.UserSlug, opts.GroupName)
	if err != nil {
		return err
	}
	project := opts.Args[0]

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPermissionsClient(client, host)
	if err != nil {
		return err
	}
	if err := pc.RevokeProjectPermission(context.Background(), project, subject); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Revoked permission from %s %s on project %s\n",
		subject.Kind, subjectLabel(subject), project)
	return nil
}

func subjectLabel(s backend.PermissionSubject) string {
	if s.Kind == "user" {
		return s.Slug
	}
	return s.Name
}
