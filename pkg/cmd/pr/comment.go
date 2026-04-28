package pr

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRComment returns the `bitbottle pr comment` parent group.
func NewCmdPRComment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "List or add general PR comments",
	}
	cmd.AddCommand(NewCmdPRCommentList(f))
	cmd.AddCommand(NewCmdPRCommentAdd(f))
	return cmd
}

func NewCmdPRCommentList(f *factory.Factory) *cobra.Command {
	var jsonFields, jqExpr, hostname string

	cmd := &cobra.Command{
		Use:   "list PR_ID",
		Short: "List general comments on a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostname)
			if err != nil {
				return err
			}
			cmts, err := client.ListPRComments(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}
			p := prCommentFields(f, jsonFields, jqExpr)
			for _, c := range cmts {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func NewCmdPRCommentAdd(f *factory.Factory) *cobra.Command {
	var body, hostname string

	cmd := &cobra.Command{
		Use:   "add PR_ID",
		Short: "Add a general comment to a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			ref, prID, client, err := resolvePRTarget(f, args, hostname)
			if err != nil {
				return err
			}
			c, err := client.AddPRComment(ref.Project, ref.Slug, prID, backend.AddPRCommentInput{Text: body})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Added comment #%d on pull request #%d\n", c.ID, prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "Comment body (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func prCommentFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.PRComment] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[backend.PRComment](f.IOStreams.Out, isTTY, jsonFields, jqExpr)
	p.AddField(format.Field[backend.PRComment]{Name: "id", Header: "ID", Extract: func(c backend.PRComment) any { return c.ID }})
	p.AddField(format.Field[backend.PRComment]{Name: "author", Header: "AUTHOR", Extract: func(c backend.PRComment) any {
		if c.Author.Slug != "" {
			return c.Author.Slug
		}
		return c.Author.DisplayName
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "createdAt", Header: "CREATED", Extract: func(c backend.PRComment) any {
		if jsonFields != "" || !isTTY {
			return c.CreatedAt.Format(time.RFC3339)
		}
		return c.CreatedAt.Format("2006-01-02 15:04")
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "text", Header: "TEXT", Extract: func(c backend.PRComment) any {
		if isTTY && len(c.Text) > 80 {
			return c.Text[:80]
		}
		return c.Text
	}})
	return p
}
