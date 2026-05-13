package revoke

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/shared"
)

// Options holds parsed flags for `perms repo revoke`.
type Options struct {
	Hostname  string
	UserSlug  string
	GroupName string
	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdRevoke builds the `perms repo revoke` cobra command.
func NewCmdRevoke(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "revoke PROJECT/REPO",
		Short: "Revoke a permission on a repository",
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
	project, slug, err := parseProjectRepo(opts.Args[0])
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
	pc, err := backend.AsPermissionsClient(client, host)
	if err != nil {
		return err
	}
	if err := pc.RevokeRepoPermission(context.Background(), project, slug, subject); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Revoked permission from %s %s on %s/%s\n",
		subject.Kind, subjectLabel(subject), project, slug)
	return nil
}

func parseProjectRepo(s string) (project, slug string, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid argument %q: expected PROJECT/REPO", s)
	}
	return parts[0], parts[1], nil
}

func subjectLabel(s backend.PermissionSubject) string {
	if s.Kind == "user" {
		return s.Slug
	}
	return s.Name
}
