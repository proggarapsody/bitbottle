package grant

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/shared"
)

// Options holds parsed flags for `perms repo grant`.
type Options struct {
	Hostname  string
	UserSlug  string
	GroupName string
	Confirm   bool
	// Args[0] = PROJECT/REPO, Args[1] = PERM
	Args []string
}

// NewCmdGrant builds the `perms repo grant` cobra command.
func NewCmdGrant(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "grant PROJECT/REPO PERM",
		Short: "Grant a permission on a repository",
		Long: `Grant a user or group a permission on a Bitbucket Server repository.

PERM must be one of: REPO_READ, REPO_WRITE, REPO_ADMIN.

If the subject already has a higher permission, a downgrade warning is printed
and confirmation is requested (pass --confirm to skip the prompt in non-interactive mode).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return grantRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&opts.UserSlug, "user", "", "User slug to grant the permission to")
	cmd.Flags().StringVar(&opts.GroupName, "group", "", "Group name to grant the permission to")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip downgrade confirmation prompt")
	return cmd
}

func grantRun(f *factory.Factory, opts *Options) error {
	subject, err := shared.SubjectFromFlags(opts.UserSlug, opts.GroupName)
	if err != nil {
		return err
	}
	project, slug, err := parseProjectRepo(opts.Args[0])
	if err != nil {
		return err
	}
	perm := opts.Args[1]

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

	// Downgrade warning.
	if !opts.Confirm {
		existing, listErr := pc.ListRepoPermissions(context.Background(), project, slug)
		if listErr == nil {
			if current, found := findGrant(existing, subject); found {
				if isDowngrade(current, perm) {
					if !f.IOStreams.IsStdoutTTY() {
						return fmt.Errorf("downgrade detected; pass --confirm to proceed")
					}
					fmt.Fprintf(f.IOStreams.Out,
						"Warning: %s %s currently has %s; granting %s will downgrade. Continue? [y/N]: ",
						subject.Kind, subjectLabel(subject), current, perm,
					)
					reader := bufio.NewReader(f.IOStreams.In)
					answer, _ := reader.ReadString('\n')
					answer = strings.TrimSpace(answer)
					if answer != "y" && answer != "Y" {
						fmt.Fprintln(f.IOStreams.Out, "Aborted.")
						return nil
					}
				}
			}
		}
	}

	if err := pc.GrantRepoPermission(context.Background(), project, slug, subject, perm); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Granted %s to %s %s on %s/%s\n",
		perm, subject.Kind, subjectLabel(subject), project, slug)
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

func findGrant(grants []backend.PermissionGrant, subject backend.PermissionSubject) (string, bool) {
	for _, g := range grants {
		if g.Subject.Kind != subject.Kind {
			continue
		}
		if subject.Kind == "user" && g.Subject.Slug == subject.Slug {
			return g.Permission, true
		}
		if subject.Kind == "group" && g.Subject.Name == subject.Name {
			return g.Permission, true
		}
	}
	return "", false
}

func permRank(p string) int {
	switch p {
	case "PROJECT_ADMIN", "REPO_ADMIN":
		return 3
	case "PROJECT_WRITE", "REPO_WRITE":
		return 2
	case "PROJECT_READ", "REPO_READ":
		return 1
	}
	return 0
}

func isDowngrade(current, newPerm string) bool {
	return permRank(newPerm) < permRank(current)
}
