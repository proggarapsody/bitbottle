package unready

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	prshared "github.com/proggarapsody/bitbottle/pkg/cmd/pr/shared"
)

func NewCmdUnready(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:   "unready PR_ID",
		Short: "Convert a pull request back to draft",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := prshared.ResolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			if err := client.UnreadyPR(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Converted pull request #%d back to draft\n", prID)
			if pr, err := client.GetPR(ref.Project, ref.Slug, prID); err == nil && pr.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", pr.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
