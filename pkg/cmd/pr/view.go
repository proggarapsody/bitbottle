package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRView(f *factory.Factory) *cobra.Command {
	var web bool
	var jsonFields string
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "view PR_ID",
		Short: "View a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			p, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			if web {
				if p.WebURL == "" {
					return fmt.Errorf("no web URL available for this pull request")
				}
				return f.Browser.Browse(p.WebURL)
			}

			if jsonFields != "" || jqExpr != "" {
				printer := prFieldsWithDescription(f, jsonFields, jqExpr)
				printer.SetSingleItem()
				printer.AddItem(p)
				return printer.Render()
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "#%d %s\n", p.ID, p.Title)
			state := p.State
			if p.Draft {
				state += " (draft)"
			}
			fmt.Fprintf(out, "State:  %s\n", state)
			if p.Description != "" {
				fmt.Fprintf(out, "\n%s\n\n", p.Description)
			}
			author := p.Author.Slug
			if p.Author.DisplayName != "" {
				author = p.Author.DisplayName
			}
			fmt.Fprintf(out, "Author: %s\n", author)
			fmt.Fprintf(out, "From:   %s\n", p.FromBranch)
			fmt.Fprintf(out, "To:     %s\n", p.ToBranch)
			if p.WebURL != "" {
				fmt.Fprintf(out, "URL:    %s\n", p.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	return cmd
}
