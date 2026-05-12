package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdPRView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view PR_ID",
		Short: "View a pull request",
		Args:  cobra.ExactArgs(1),
		// Long PR descriptions plus reviewer/build-status sections often
		// exceed a terminal page; route through $PAGER on a TTY. The
		// annotation handler at the root wires StartPager/StopPager.
		Annotations: map[string]string{cmdutil.PagerAnnotation: "true"},
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

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				printer := prFieldsWithDescription(f, cfg)
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
			if p.AutoMerge != nil && p.AutoMerge.Enabled {
				fmt.Fprintf(out, "Auto-merge: enabled (%s)\n", p.AutoMerge.Strategy)
			}
			if p.WebURL != "" {
				fmt.Fprintf(out, "URL:    %s\n", p.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	return cmd
}
