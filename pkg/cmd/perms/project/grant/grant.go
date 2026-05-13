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

// Options holds parsed flags for `perms project grant`.
type Options struct {
	Hostname  string
	UserSlug  string
	GroupName string
	Force     bool
	// Args[0] = PROJECT, Args[1] = PERM
	Args []string
}

// NewCmdGrant builds the `perms project grant` cobra command.
func NewCmdGrant(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "grant PROJECT PERM",
		Short: "Grant a permission on a project",
		Long: `Grant a user or group a permission on a Bitbucket Server project.

PERM must be one of: PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN.

If the subject already has a higher permission, a downgrade warning is printed
and confirmation is requested (--force or non-TTY skips the prompt).`,
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
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip downgrade confirmation prompt")
	return cmd
}

func grantRun(f *factory.Factory, opts *Options) error {
	subject, err := shared.SubjectFromFlags(opts.UserSlug, opts.GroupName)
	if err != nil {
		return err
	}
	project := opts.Args[0]
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

	// Downgrade warning: fetch existing grants to detect if we're lowering the level.
	if !opts.Force && f.IOStreams.IsStdoutTTY() {
		existing, listErr := pc.ListProjectPermissions(context.Background(), project)
		if listErr == nil {
			if current, found := findGrant(existing, subject); found {
				if isDowngrade(current, perm) {
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

	if err := pc.GrantProjectPermission(context.Background(), project, subject, perm); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Granted %s to %s %s on project %s\n",
		perm, subject.Kind, subjectLabel(subject), project)
	return nil
}

func subjectLabel(s backend.PermissionSubject) string {
	if s.Kind == "user" {
		return s.Slug
	}
	return s.Name
}

// findGrant returns the permission currently held by subject in grants.
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

// permRank returns a numeric ordering for permission levels.
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

// isDowngrade returns true when newPerm is a lower privilege than current.
func isDowngrade(current, newPerm string) bool {
	return permRank(newPerm) < permRank(current)
}
