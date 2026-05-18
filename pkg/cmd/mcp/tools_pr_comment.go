package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRCommentTools)
}

func registerPRCommentTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key or workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	reqID := mcplib.WithNumber("id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
	)
	reqCommentID := mcplib.WithNumber("comment_id",
		mcplib.Description("Comment ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("submit_pr_review",
			mcplib.WithDescription("Submit a compound pull-request review (top-level body + inline comments + an action). Action defaults to \"comment\" when body or inline_comments is set without an explicit action; \"request_changes\" is Bitbucket Cloud only and surfaces host.unsupported on Server."),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithString("action",
				mcplib.Description(`Review action: "approve", "request_changes", or "comment" (default)`),
			),
			mcplib.WithString("body",
				mcplib.Description("Top-level review body comment"),
			),
			mcplib.WithArray("inline_comments",
				mcplib.Description("Inline review comments anchored to file:line in the diff. Each item: {path, line, body}; optional start_line for ranges (Cloud only); optional side (\"new\" or \"old\", default \"new\")."),
				mcplib.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path":       map[string]any{"type": "string"},
						"line":       map[string]any{"type": "number"},
						"body":       map[string]any{"type": "string"},
						"start_line": map[string]any{"type": "number"},
						"side":       map[string]any{"type": "string", "enum": []string{"new", "old"}},
					},
					"required": []string{"path", "line", "body"},
				}),
			),
		),
		h.submitPRReview,
	)

	s.AddTool(
		mcplib.NewTool("list_pr_comments",
			mcplib.WithDescription("List comments on a pull request, including inline (file:line) review comments and replies"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithBoolean("inline_only",
				mcplib.Description("Only return inline (file:line) review comments (default: false)"),
			),
			mcplib.WithBoolean("include_reactions",
				mcplib.Description("Fetch and include emoji reactions for each comment (Bitbucket Server / DC only; default: false)"),
			),
		),
		h.listPRComments,
	)

	s.AddTool(
		mcplib.NewTool("add_pr_comment",
			mcplib.WithDescription("Add a comment to a pull request. Optionally anchor to a file:line as an inline review comment, or reply nested under an existing comment."),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithString("body",
				mcplib.Description("Comment body"),
				mcplib.Required(),
			),
			mcplib.WithString("inline_path",
				mcplib.Description("File path in the PR diff to anchor an inline comment to (omit for general comment)"),
			),
			mcplib.WithNumber("inline_line",
				mcplib.Description("Line number in the diff (required when inline_path is set)"),
			),
			mcplib.WithNumber("inline_start_line",
				mcplib.Description("First line of a multi-line range (Cloud only); leave 0 for single-line"),
			),
			mcplib.WithString("inline_side",
				mcplib.Description(`Diff side for the inline anchor: "new" (default) or "old"`),
			),
			mcplib.WithNumber("parent_id",
				mcplib.Description("ID of the comment to reply under (omit for a top-level comment)"),
			),
		),
		h.addPRComment,
	)

	s.AddTool(
		mcplib.NewTool("edit_pr_comment",
			mcplib.WithDescription("Update the body of an existing comment on a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			reqCommentID,
			mcplib.WithString("body",
				mcplib.Description("New comment body"),
				mcplib.Required(),
			),
		),
		h.editPRComment,
	)

	s.AddTool(
		mcplib.NewTool("delete_pr_comment",
			mcplib.WithDescription("Delete an existing comment from a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			reqCommentID,
		),
		h.deletePRComment,
	)

	s.AddTool(
		mcplib.NewTool("resolve_pr_comment",
			mcplib.WithDescription("Mark a pull-request comment thread as resolved (Bitbucket Cloud only; Server returns host.unsupported)"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			reqCommentID,
		),
		h.resolvePRComment,
	)
}
