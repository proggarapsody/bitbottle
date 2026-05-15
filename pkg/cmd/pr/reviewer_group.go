package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdReviewerGroup builds the `pr reviewer-group` parent command.
func NewCmdReviewerGroup(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviewer-group",
		Short: "Manage PR reviewer groups for a repository",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory. Bitbucket Server / Data Center only.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdReviewerGroupList(f))
	cmd.AddCommand(NewCmdReviewerGroupAdd(f))
	cmd.AddCommand(NewCmdReviewerGroupRemove(f))
	return cmd
}

// NewCmdReviewerGroupList builds the `pr reviewer-group list` command.
func NewCmdReviewerGroupList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List reviewer groups for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rg, err := backend.AsReviewerGroupClient(client, ref.Host)
			if err != nil {
				return err
			}
			groups, err := rg.ListReviewerGroups(ref.Project, ref.Slug)
			if err != nil {
				return err
			}
			p := reviewerGroupFields(f, format.ConfigFromCmd(cmd))
			for _, g := range groups {
				p.AddItem(g)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

// NewCmdReviewerGroupAdd builds the `pr reviewer-group add` command.
func NewCmdReviewerGroupAdd(f *factory.Factory) *cobra.Command {
	var (
		hostname          string
		name              string
		users             string
		requiredApprovals int
	)
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO]",
		Short: "Create a reviewer group for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if users == "" {
				return fmt.Errorf("--users is required")
			}
			userSlugs := splitUsers(users)

			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rg, err := backend.AsReviewerGroupClient(client, ref.Host)
			if err != nil {
				return err
			}
			if requiredApprovals <= 0 {
				requiredApprovals = 1
			}
			group, err := rg.CreateReviewerGroup(ref.Project, ref.Slug, backend.CreateReviewerGroupInput{
				Name:              name,
				UserSlugs:         userSlugs,
				RequiredApprovals: requiredApprovals,
			})
			if err != nil {
				return err
			}
			p := reviewerGroupFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(group)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&name, "name", "", "Name for the reviewer group (required)")
	cmd.Flags().StringVar(&users, "users", "", "Comma-separated list of user slugs (required)")
	cmd.Flags().IntVar(&requiredApprovals, "required-approvals", 1, "Required number of approvals")
	format.RegisterOutputFlags(cmd)
	return cmd
}

// NewCmdReviewerGroupRemove builds the `pr reviewer-group remove` command.
func NewCmdReviewerGroupRemove(f *factory.Factory) *cobra.Command {
	var (
		hostname string
		idFlag   int
	)
	cmd := &cobra.Command{
		Use:   "remove [PROJECT/REPO] NAME",
		Short: "Remove a reviewer group from a repository",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine repo ref and name/id from args.
			// Accepted forms:
			//   remove NAME            (repo inferred from git remote)
			//   remove PROJECT/REPO NAME
			//   remove [PROJECT/REPO] --id ID
			var repoArgs []string
			var groupName string

			if idFlag == 0 {
				// name mode
				switch len(args) {
				case 0:
					return fmt.Errorf("NAME argument required (or use --id)")
				case 1:
					// Could be PROJECT/REPO or NAME — treat as NAME when it
					// contains no slash, else treat as repo (and still need name)
					if !strings.Contains(args[0], "/") {
						groupName = args[0]
					} else {
						return fmt.Errorf("NAME argument required when PROJECT/REPO is specified")
					}
				case 2:
					repoArgs = args[:1]
					groupName = args[1]
				}
			} else {
				// id mode: at most one positional arg (the repo)
				if len(args) > 1 {
					return fmt.Errorf("too many arguments")
				}
				repoArgs = args
			}

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rg, err := backend.AsReviewerGroupClient(client, ref.Host)
			if err != nil {
				return err
			}

			deleteID := idFlag
			if deleteID == 0 {
				// Look up by name
				groups, err := rg.ListReviewerGroups(ref.Project, ref.Slug)
				if err != nil {
					return err
				}
				for _, g := range groups {
					if g.Name == groupName {
						deleteID = g.ID
						break
					}
				}
				if deleteID == 0 {
					return fmt.Errorf("reviewer group %q not found", groupName)
				}
			}

			if err := rg.DeleteReviewerGroup(ref.Project, ref.Slug, deleteID); err != nil {
				return err
			}
			if idFlag != 0 {
				fmt.Fprintf(f.IOStreams.Out, "Reviewer group %d removed\n", deleteID)
			} else {
				fmt.Fprintf(f.IOStreams.Out, "Reviewer group %q removed\n", groupName)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().IntVar(&idFlag, "id", 0, "Condition ID to remove directly (skips name lookup)")
	return cmd
}

func reviewerGroupFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.ReviewerGroup] {
	p := format.New[backend.ReviewerGroup](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.ReviewerGroup]{
		Name:    "id",
		Header:  "ID",
		Extract: func(g backend.ReviewerGroup) any { return g.ID },
	})
	p.AddField(format.Field[backend.ReviewerGroup]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(g backend.ReviewerGroup) any { return g.Name },
	})
	p.AddField(format.Field[backend.ReviewerGroup]{
		Name:    "requiredApprovals",
		Header:  "REQUIRED APPROVALS",
		Extract: func(g backend.ReviewerGroup) any { return g.RequiredApprovals },
	})
	p.AddField(format.Field[backend.ReviewerGroup]{
		Name:   "reviewers",
		Header: "REVIEWERS",
		Extract: func(g backend.ReviewerGroup) any {
			slugs := make([]string, 0, len(g.Reviewers))
			for _, r := range g.Reviewers {
				slugs = append(slugs, r.Slug)
			}
			return strings.Join(slugs, ", ")
		},
	})
	return p
}

// splitUsers splits a comma-separated list of user slugs, trimming whitespace.
func splitUsers(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
