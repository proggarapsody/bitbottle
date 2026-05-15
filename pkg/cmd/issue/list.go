package issue

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// mapIssueState normalises CLI-friendly state names to Bitbucket Cloud's
// lowercase API values. "all" means no filter. We mirror gh's mapPRState
// pattern so users can type --state open / closed and it Just Works.
func mapIssueState(state string) string {
	switch strings.ToLower(state) {
	case "all", "":
		return ""
	case "closed":
		return "closed"
	case "open":
		return "open"
	case "new":
		return "new"
	case "on-hold", "on_hold":
		return "on hold"
	default:
		return strings.ToLower(state)
	}
}

func NewCmdIssueList(f *factory.Factory) *cobra.Command {
	var state, hostname string
	var limit int
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List issues in a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			ic, err := backend.AsIssueClient(client, ref.Host)
			if err != nil {
				return err
			}
			issues, err := ic.ListIssues(ref.Project, ref.Slug, mapIssueState(state), limit)
			if err != nil {
				return err
			}
			p := issueListFields(f, format.ConfigFromCmd(cmd))
			for _, i := range issues {
				p.AddItem(i)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "State filter: open, new, on-hold, resolved, duplicate, invalid, wontfix, closed, all")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of issues")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
