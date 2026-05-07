package protect

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// CreateOptions holds parsed flags for `branch protect create`.
type CreateOptions struct {
	Hostname string
	Type     string
	Branch   string
	Pattern  string
	Users    []string
	Groups   []string

	// Args[0] = PROJECT/REPO
	Args []string
}

// validTypes lists the four restriction types Bitbucket Server / DC accepts.
// Validation happens client-side so users get a clear error before we send a
// request the server is going to reject anyway.
var validTypes = []string{"read-only", "no-deletes", "fast-forward-only", "pull-request-only"}

// NewCmdCreate builds the `branch protect create` cobra command.
func NewCmdCreate(f *factory.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{}
	cmd := &cobra.Command{
		Use:   "create PROJECT/REPO",
		Short: "Add a branch restriction",
		Long: `Add a branch restriction to a Bitbucket Server / DC repository.

Specify exactly one of --branch (a literal branch name) or --pattern (a glob
like "release/*"). The --type flag chooses the restriction kind:

  read-only          disallow all writes
  no-deletes         disallow branch deletion
  fast-forward-only  disallow non-fast-forward writes
  pull-request-only  only PR merges may write

Pass --user (repeatable) and --group (repeatable) to exempt specific users
or groups from the restriction.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Type, "type", "", fmt.Sprintf("Restriction type (one of: %s)", strings.Join(validTypes, ", ")))
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Restrict a single branch by name")
	cmd.Flags().StringVar(&opts.Pattern, "pattern", "", "Restrict branches matching a glob pattern (e.g. \"release/*\")")
	cmd.Flags().StringSliceVar(&opts.Users, "user", nil, "User slug to exempt from the restriction (repeatable)")
	cmd.Flags().StringSliceVar(&opts.Groups, "group", nil, "Group slug to exempt from the restriction (repeatable)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func createRun(f *factory.Factory, opts *CreateOptions) error {
	if !contains(validTypes, opts.Type) {
		return fmt.Errorf("--type %q: must be one of %s", opts.Type, strings.Join(validTypes, ", "))
	}
	if (opts.Branch == "") == (opts.Pattern == "") {
		return fmt.Errorf("specify exactly one of --branch or --pattern")
	}
	matcherID := opts.Branch
	matcherKind := "BRANCH"
	if opts.Pattern != "" {
		matcherID = opts.Pattern
		matcherKind = "PATTERN"
	}

	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	bp, err := backend.AsBranchProtector(client, ref.Host)
	if err != nil {
		return err
	}
	out, err := bp.CreateBranchProtection(ref.Project, ref.Slug, backend.CreateBranchProtectionInput{
		Type:        opts.Type,
		MatcherID:   matcherID,
		MatcherKind: matcherKind,
		Users:       opts.Users,
		Groups:      opts.Groups,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Created restriction %d: %s on %s\n", out.ID, out.Type, out.MatcherID)
	return nil
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
