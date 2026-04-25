package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRMerge(f *factory.Factory) *cobra.Command {
	var merge, squash, deleteBranch bool

	cmd := &cobra.Command{
		Use:   "merge PR_ID",
		Short: "Merge a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if merge && squash {
				return fmt.Errorf("cannot use --merge and --squash together")
			}

			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			var strategy string
			switch {
			case merge:
				strategy = "merge-commit"
			case squash:
				strategy = "squash"
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
	cmd.Flags().BoolVar(&deleteBranch, "delete-branch", false, "Delete source branch after merge")
	return cmd
}
