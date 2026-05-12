package issue

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssueCreate(f *factory.Factory) *cobra.Command {
	var title, body, kind, priority, hostname string
	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO]",
		Short: "Create a new issue",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			issue, err := ic.CreateIssue(ref.Project, ref.Slug, backend.CreateIssueInput{
				Title:    title,
				Content:  body,
				Kind:     kind,
				Priority: priority,
			})
			if err != nil {
				return err
			}
			p := issueViewFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(issue)
			return p.Render()
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "Issue title (required)")
	cmd.Flags().StringVarP(&body, "body", "b", "", "Issue body (markdown)")
	cmd.Flags().StringVar(&kind, "kind", "", "Issue kind: bug, enhancement, proposal, task")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority: trivial, minor, major, critical, blocker")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
