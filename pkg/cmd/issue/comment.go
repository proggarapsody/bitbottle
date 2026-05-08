package issue

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdIssueComment returns the `issue comment` subcommand group.
func NewCmdIssueComment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage comments on a Bitbucket Cloud issue",
	}
	cmd.AddCommand(NewCmdIssueCommentList(f))
	cmd.AddCommand(NewCmdIssueCommentAdd(f))
	cmd.AddCommand(NewCmdIssueCommentEdit(f))
	cmd.AddCommand(NewCmdIssueCommentDelete(f))
	return cmd
}

func issueCommentFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.IssueComment] {
	p := format.New[backend.IssueComment](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.IssueComment]{Name: "id", Header: "ID", Extract: func(c backend.IssueComment) any { return c.ID }})
	p.AddField(format.Field[backend.IssueComment]{Name: "author", Header: "AUTHOR", Extract: func(c backend.IssueComment) any { return c.Author.Slug }})
	p.AddField(format.Field[backend.IssueComment]{Name: "content", Header: "CONTENT", Extract: func(c backend.IssueComment) any { return c.Content }})
	p.AddField(format.Field[backend.IssueComment]{Name: "createdOn", Header: "CREATED", Extract: func(c backend.IssueComment) any {
		if c.CreatedOn.IsZero() {
			return ""
		}
		return c.CreatedOn.Format(time.RFC3339)
	}})
	return p
}

func NewCmdIssueCommentList(f *factory.Factory) *cobra.Command {
	var jsonFields, jqExpr, hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO] ISSUE_ID",
		Short: "List comments on an issue",
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
			comments, err := ic.ListIssueComments(ref.Project, ref.Slug, id)
			if err != nil {
				return err
			}
			p := issueCommentFields(f, jsonFields, jqExpr)
			for _, c := range comments {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func NewCmdIssueCommentAdd(f *factory.Factory) *cobra.Command {
	var hostname, body string
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO] ISSUE_ID",
		Short: "Add a comment to an issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
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
			comment, err := ic.AddIssueComment(ref.Project, ref.Slug, id, body)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Added comment #%d to issue #%d\n", comment.ID, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&body, "body", "", "Comment body (required)")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

// splitCommentIDArgs handles: [PROJECT/REPO] ISSUE_ID COMMENT_ID
// Returns (repoArgs, issueID, commentID).
func splitCommentIDArgs(args []string) ([]string, string, string) {
	if len(args) == 3 {
		return []string{args[0]}, args[1], args[2]
	}
	return nil, args[0], args[1]
}

func NewCmdIssueCommentEdit(f *factory.Factory) *cobra.Command {
	var hostname, body string
	cmd := &cobra.Command{
		Use:   "edit [PROJECT/REPO] ISSUE_ID COMMENT_ID",
		Short: "Edit a comment on an issue",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			repoArgs, idArg, commentIDArg := splitCommentIDArgs(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			commentID, err := strconv.Atoi(commentIDArg)
			if err != nil {
				return fmt.Errorf("invalid comment ID %q: must be a number", commentIDArg)
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
			comment, err := ic.EditIssueComment(ref.Project, ref.Slug, id, commentID, body)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Updated comment #%d on issue #%d\n", comment.ID, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&body, "body", "", "New comment body (required)")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func NewCmdIssueCommentDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] ISSUE_ID COMMENT_ID",
		Short: "Delete a comment on an issue",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg, commentIDArg := splitCommentIDArgs(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			commentID, err := strconv.Atoi(commentIDArg)
			if err != nil {
				return fmt.Errorf("invalid comment ID %q: must be a number", commentIDArg)
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
			if err := ic.DeleteIssueComment(ref.Project, ref.Slug, id, commentID); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted comment #%d from issue #%d\n", commentID, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
