package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRReady(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:   "ready PR_ID",
		Short: "Mark a pull request as ready for review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			if err := client.ReadyPR(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Marked pull request #%d as ready for review\n", prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
