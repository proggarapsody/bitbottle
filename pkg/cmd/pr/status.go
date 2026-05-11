package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRStatus returns the `bb pr status` command.
// It shows open PRs authored by and assigned to the current user.
func NewCmdPRStatus(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "status [PROJECT/REPO]",
		Short: "Show open pull requests authored by or assigned to you",
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

			lister, ok := client.(backend.PRStatusLister)
			if !ok {
				return fmt.Errorf("pr status is not supported by this backend")
			}

			entries, err := lister.ListMyPRs(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			out := f.IOStreams.Out
			printSection(out, "Pull requests assigned to you for review", entries, "REVIEWER")
			printSection(out, "Pull requests created by you", entries, "AUTHOR")
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func printSection(out interface{ Write([]byte) (int, error) }, heading string, entries []backend.MyPREntry, role string) {
	fmt.Fprintf(out, "\n%s\n\n", heading)
	fmt.Fprintf(out, "%-6s %-40s %-25s %s\n", "PR#", "TITLE", "REPO", "STATE")
	found := false
	for _, e := range entries {
		if e.Role != role {
			continue
		}
		found = true
		title := e.Title
		if len(title) > 38 {
			title = title[:35] + "..."
		}
		fmt.Fprintf(out, "#%-5d %-40s %-25s %s\n", e.ID, title, e.Repo, e.State)
	}
	if !found {
		fmt.Fprintf(out, "  (none)\n")
	}
}

// resolveRepoRef is a helper for commands that need a RepoRef but not a PR ID.
func resolveRepoRef(f *factory.Factory, args []string, hostname string) (bbrepo.RepoRef, backend.Client, error) {
	ref, err := factory.ResolveTarget(f, args, hostname)
	if err != nil {
		return bbrepo.RepoRef{}, nil, err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return bbrepo.RepoRef{}, nil, err
	}
	return ref, client, nil
}
