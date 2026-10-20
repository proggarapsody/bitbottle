package pr

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRMerge(f *factory.Factory) *cobra.Command {
	var merge, squash, rebase, deleteBranch, auto, autoOff bool

	cmd := &cobra.Command{
		Use:   "merge PR_ID",
		Short: "Merge a pull request or request automatic merge",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (merge && squash) || (merge && rebase) || (squash && rebase) {
				return fmt.Errorf("choose only one of --merge, --squash, or --rebase")
			}
			if auto && autoOff {
				return fmt.Errorf("cannot use --auto and --auto-off together")
			}

			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			// --auto-off: cancel a queued auto-merge.
			if autoOff {
				if err := client.DisableAutoMerge(ref.Project, ref.Slug, prID); err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "Cancelled auto-merge for pull request #%d\n", prID)
				return nil
			}

			// --auto: queue for auto-merge when checks pass.
			if auto {
				var strategy string
				switch {
				case squash:
					strategy = "squash"
				case rebase:
					strategy = "rebase"
				default:
					strategy = "merge"
				}
				if err := client.EnableAutoMerge(ref.Project, ref.Slug, prID, strategy); err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "Auto-merge requested for pull request #%d (%s)\n", prID, strategy)
				return nil
			}

			// Normal merge path. If the PR has auto-merge queued, prompt before
			// overriding it (race-condition guard).
			current, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}
			if current.AutoMerge != nil && current.AutoMerge.Enabled {
				if f.IOStreams.IsStdoutTTY() {
					fmt.Fprint(f.IOStreams.Out, "PR is queued for auto-merge. Cancel and merge now? [y/N] ")
					reader := bufio.NewReader(f.IOStreams.In)
					answer, _ := reader.ReadString('\n')
					answer = strings.TrimSpace(strings.ToLower(answer))
					if answer != "y" {
						fmt.Fprintln(f.IOStreams.Out, "Aborted.")
						return nil
					}
				}
				// Cancel the queued auto-merge before the manual merge.
				if err := client.DisableAutoMerge(ref.Project, ref.Slug, prID); err != nil {
					return fmt.Errorf("failed to cancel auto-merge: %w", err)
				}
			}

			var strategy string
			switch {
			case merge:
				strategy = "merge-commit"
			case squash:
				strategy = "squash"
			case rebase:
				strategy = "rebase"
			default:
				strategy = ""
			}

			pr, err := client.MergePR(ref.Project, ref.Slug, prID, backend.MergePRInput{
				Strategy: strategy,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Merged pull request #%d\n", prID)

			if deleteBranch {
				if err := client.DeleteBranch(ref.Project, ref.Slug, pr.FromBranch); err != nil {
					return fmt.Errorf("merge succeeded but failed to delete branch: %w", err)
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge commit strategy")
	cmd.Flags().BoolVar(&squash, "squash", false, "Squash merge strategy")
	cmd.Flags().BoolVar(&rebase, "rebase", false, "Rebase and fast-forward merge strategy")
	cmd.Flags().BoolVar(&deleteBranch, "delete-branch", false, "Delete source branch after merge")
	cmd.Flags().BoolVar(&auto, "auto", false, "Request auto-merge (Cloud: active queue)")
	cmd.Flags().BoolVar(&autoOff, "auto-off", false, "Cancel auto-merge (Cloud: remove from queue)")
	return cmd
}
