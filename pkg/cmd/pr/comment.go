package pr

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

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
	var inlineOnly bool

	cmd := &cobra.Command{
		Use:   "list PR_ID",
		Short: "List comments on a pull request (general + inline review comments)",
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
			if inlineOnly {
				cmts = filterInlinePRComments(cmts)
			}
			p := prCommentFields(f, jsonFields, jqExpr, hasInline(cmts))
			for _, c := range cmts {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().BoolVar(&inlineOnly, "inline", false, "Only show inline (file:line) review comments")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func filterInlinePRComments(cmts []backend.PRComment) []backend.PRComment {
	out := make([]backend.PRComment, 0, len(cmts))
	for _, c := range cmts {
		if c.Inline != nil {
			out = append(out, c)
		}
	}
	return out
}

func hasInline(cmts []backend.PRComment) bool {
	for _, c := range cmts {
		if c.Inline != nil {
			return true
		}
	}
	return false
}

// formatInlineLocation renders an inline anchor as "path:line" or
// "path:start-end" for multi-line comments. Returns "" for nil.
func formatInlineLocation(in *backend.PRCommentInline) string {
	if in == nil {
		return ""
	}
	if in.StartLine != 0 && in.StartLine != in.Line {
		return fmt.Sprintf("%s:%d-%d", in.Path, in.StartLine, in.Line)
	}
	return fmt.Sprintf("%s:%d", in.Path, in.Line)
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

func prCommentFields(f *factory.Factory, jsonFields, jqExpr string, showLocation bool) *format.Printer[backend.PRComment] {
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
	if showLocation {
		p.AddField(format.Field[backend.PRComment]{Name: "location", Header: "LOCATION", Extract: func(c backend.PRComment) any {
			return formatInlineLocation(c.Inline)
		}})
	}
	p.AddField(format.Field[backend.PRComment]{Name: "text", Header: "TEXT", Extract: func(c backend.PRComment) any {
		if isTTY && len(c.Text) > 80 {
			return c.Text[:80]
		}
		return c.Text
	}})
	// JSON-only fields: surfaced via --json but not in the default TTY table.
	p.AddField(format.Field[backend.PRComment]{Name: "inline", Header: "INLINE", JSONOnly: true, Extract: func(c backend.PRComment) any {
		if c.Inline == nil {
			return nil
		}
		m := map[string]any{"path": c.Inline.Path, "side": c.Inline.Side, "line": c.Inline.Line}
		if c.Inline.StartLine != 0 {
			m["startLine"] = c.Inline.StartLine
		}
		return m
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "parentId", Header: "PARENT", JSONOnly: true, Extract: func(c backend.PRComment) any {
		return c.ParentID
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "updatedAt", Header: "UPDATED", JSONOnly: true, Extract: func(c backend.PRComment) any {
		if c.UpdatedAt.IsZero() {
			return nil
		}
		return c.UpdatedAt.Format(time.RFC3339)
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "resolved", Header: "RESOLVED", JSONOnly: true, Extract: func(c backend.PRComment) any {
		return c.Resolved
	}})
	return p
}
