package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRDiff(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff PR_ID",
		Short: "Show a pull request diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			diff, err := client.GetPRDiff(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			fmt.Fprint(f.IOStreams.Out, diff)
			return nil
		},
	}
	return cmd
}
