package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRCreate(f *factory.Factory) *cobra.Command {
	var title, body, base, head string
	var draft, noDefaultReviewers bool
	var reviewers []string
	var jsonFields string
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, "")
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)

			currentBranch := head
			if currentBranch == "" {
				g := git.New(f.GitRunner())
				currentBranch, err = g.CurrentBranch()
				if err != nil {
					return fmt.Errorf("%w\nhint: use --head to specify the source branch explicitly", err)
				}
			}

			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if base == "" {
				return fmt.Errorf("--base is required")
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			finalReviewers := mergeReviewers(reviewers, fetchDefaultReviewers(
				f, client, ref.Host, ref.Project, ref.Slug, currentBranch, base, noDefaultReviewers,
			))

			p, err := client.CreatePR(ref.Project, ref.Slug, backend.CreatePRInput{
				Title:       title,
				Description: body,
				Draft:       draft,
				FromBranch:  currentBranch,
				ToBranch:    base,
				Reviewers:   finalReviewers,
			})
			if err != nil {
				return err
			}

			if jsonFields != "" || jqExpr != "" {
				printer := prFields(f, jsonFields, jqExpr)
				printer.SetSingleItem()
				printer.AddItem(p)
				return printer.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Created pull request #%d: %s\n", p.ID, p.Title)
			if p.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", p.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Pull request title")
	cmd.Flags().StringVar(&body, "body", "", "Pull request description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch")
	cmd.Flags().StringVar(&head, "head", "", "Source branch (default: current branch from git)")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft")
	cmd.Flags().StringSliceVar(&reviewers, "reviewer", nil, "Reviewer user slug (repeatable; combines with auto-applied default reviewers)")
	cmd.Flags().BoolVar(&noDefaultReviewers, "no-default-reviewers", false, "Skip auto-applying the repo's configured default reviewers")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	return cmd
}

// fetchDefaultReviewers returns the user slugs configured as default reviewers
// for this PR, or nil when:
//   - the user passed --no-default-reviewers,
//   - the backend doesn't support default reviewers (Bitbucket Cloud),
//   - the lookup fails (we degrade gracefully — a warning is printed and PR
//     creation proceeds with whatever explicit --reviewer flags supplied).
//
// This matches the Bitbucket Server web UI behaviour, where default reviewers
// are applied automatically. Bitbucket's REST API does NOT auto-apply them on
// the create-PR endpoint, so the client has to do the merge.
func fetchDefaultReviewers(
	f *factory.Factory, client backend.Client, host, project, slug, fromBranch, toBranch string,
	skip bool,
) []string {
	if skip {
		return nil
	}
	resolver, err := backend.AsDefaultReviewersResolver(client, host)
	if err != nil {
		return nil // Cloud — silently skip, --reviewer is the only path.
	}
	users, err := resolver.DefaultReviewers(project, slug, fromBranch, toBranch)
	if err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut,
			"warning: could not fetch default reviewers (%s); continuing without them\n", err)
		return nil
	}
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Slug)
	}
	return out
}

// mergeReviewers concatenates the explicit --reviewer flags with any
// auto-fetched defaults, deduping while preserving the explicit-first order
// (so the user's intent ranks above the repo's defaults). Empty entries
// are dropped.
func mergeReviewers(explicit, defaults []string) []string {
	seen := make(map[string]struct{}, len(explicit)+len(defaults))
	out := make([]string, 0, len(explicit)+len(defaults))
	for _, list := range [][]string{explicit, defaults} {
		for _, slug := range list {
			if slug == "" {
				continue
			}
			if _, dup := seen[slug]; dup {
				continue
			}
			seen[slug] = struct{}{}
			out = append(out, slug)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
