package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRUpdateBranch returns the `bb pr update-branch PR_ID` command.
// It syncs the PR's source branch with its base branch.
func NewCmdPRUpdateBranch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-branch PR_ID",
		Short: "Sync a PR's source branch with its base branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			updater, ok := client.(backend.PRBranchUpdater)
			if !ok {
				return fmt.Errorf("update-branch is not supported by this backend")
			}

			if err := updater.UpdatePRBranch(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintln(f.IOStreams.Out, "branch updated")
			return nil
		},
	}
	return cmd
}
