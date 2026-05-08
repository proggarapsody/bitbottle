package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssueEdit(f *factory.Factory) *cobra.Command {
	var (
		hostname string
		title    string
		body     string
		kind     string
		priority string
		assignee string
		state    string
	)
	cmd := &cobra.Command{
		Use:   "edit [PROJECT/REPO] ID",
		Short: "Edit an issue (title, body, kind, priority, assignee, state)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
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
			in := backend.UpdateIssueInput{
				Title:    title,
				Content:  body,
				Kind:     kind,
				Priority: priority,
				State:    state,
				Assignee: assignee,
			}
			issue, err := ic.UpdateIssue(ref.Project, ref.Slug, id, in)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Updated issue #%d: %s\n", issue.ID, issue.Title)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&body, "body", "", "New body (markdown)")
	cmd.Flags().StringVar(&kind, "kind", "", "Kind: bug, enhancement, proposal, task")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority: trivial, minor, major, critical, blocker")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee username (use \"\" to clear)")
	cmd.Flags().StringVar(&state, "state", "", "State: new, open, resolved, on hold, invalid, duplicate, wontfix, closed")
	return cmd
}
