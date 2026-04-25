package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view PR_ID",
		Short: "View a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			pr, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			// --web: open in browser and skip text output.
			if web {
				if pr.WebURL == "" {
					return fmt.Errorf("no web URL available for this pull request")
				}
				return f.Browser.Browse(pr.WebURL)
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "#%d %s\n", pr.ID, pr.Title)
			fmt.Fprintf(out, "State:  %s\n", pr.State)
			fmt.Fprintf(out, "Author: %s\n", pr.Author.Slug)
			fmt.Fprintf(out, "From:   %s\n", pr.FromBranch)
			fmt.Fprintf(out, "To:     %s\n", pr.ToBranch)
			if pr.WebURL != "" {
				fmt.Fprintf(out, "URL:    %s\n", pr.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	return cmd
}
