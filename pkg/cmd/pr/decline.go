package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRDecline(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:   "decline PR_ID",
		Short: "Decline a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			if err := client.DeclinePR(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Declined pull request #%d\n", prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
