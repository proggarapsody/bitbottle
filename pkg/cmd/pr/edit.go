package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPREdit(f *factory.Factory) *cobra.Command {
	var title, body, removeReviewer, hostnameFlag string

	cmd := &cobra.Command{
		Use:   "edit PR_ID",
		Short: "Edit a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" && body == "" && removeReviewer == "" {
				return fmt.Errorf("specify at least --title, --body, or --remove-reviewer")
			}

			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			pr, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}
			if err := backend.ValidateMutablePRState(pr); err != nil {
				return err
			}

			if title != "" || body != "" {
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
			}

			if removeReviewer != "" {
				var users []string
				for _, u := range strings.Split(removeReviewer, ",") {
					if u = strings.TrimSpace(u); u != "" {
						users = append(users, u)
					}
				}
				remover, ok := client.(backend.PRReviewerRemover)
				if !ok {
					return fmt.Errorf("--remove-reviewer is not supported on this host")
				}
				if err := remover.RemoveReviewers(ref.Project, ref.Slug, prID, users); err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "Removed reviewer(s) from pull request #%d\n", prID)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "New pull request title")
	cmd.Flags().StringVar(&body, "body", "", "New pull request description")
	cmd.Flags().StringVar(&removeReviewer, "remove-reviewer", "", "Comma-separated list of reviewers to remove")
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
