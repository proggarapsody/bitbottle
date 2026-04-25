package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRApprove(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve PR_ID",
		Short: "Approve a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			if err := client.ApprovePR(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Approved pull request #%d\n", prID)
			return nil
		},
	}
	return cmd
}
