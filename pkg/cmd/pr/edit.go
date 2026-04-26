package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPREdit(f *factory.Factory) *cobra.Command {
	var title, body, hostnameFlag string

	cmd := &cobra.Command{
		Use:   "edit PR_ID",
		Short: "Edit a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" && body == "" {
				return fmt.Errorf("specify at least --title or --body")
			}

			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			p, err := client.UpdatePR(ref.Project, ref.Slug, prID, backend.UpdatePRInput{
				Title:       title,
				Description: body,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Updated pull request #%d\n", p.ID)
			if p.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", p.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "New pull request title")
	cmd.Flags().StringVar(&body, "body", "", "New pull request description")
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
