package root

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdStatus returns the `bb status` top-level command.
// It shows open PRs authored by or assigned to the current user.
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "status [PROJECT/REPO]",
		Short: "Show your open pull requests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			lister, ok := client.(backend.PRStatusLister)
			if !ok {
				return fmt.Errorf("status is not supported by this backend")
			}

			entries, err := lister.ListMyPRs(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			out := f.IOStreams.Out
			printStatusSection(out, "Pull requests assigned to you for review", entries, "REVIEWER")
			printStatusSection(out, "Pull requests created by you", entries, "AUTHOR")
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func printStatusSection(out interface{ Write([]byte) (int, error) }, heading string, entries []backend.MyPREntry, role string) {
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
