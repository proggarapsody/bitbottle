package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIssueCommentTools)
}

func registerIssueCommentTools(s *mcpserver.MCPServer, h *handlers) {
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
	reqIssueID := mcplib.WithNumber("id",
		mcplib.Description("Issue ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_issue_comments",
			mcplib.WithDescription("List comments on an issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.listIssueComments,
	)

	s.AddTool(
		mcplib.NewTool("add_issue_comment",
			mcplib.WithDescription("Add a comment to an issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithString("body",
				mcplib.Description("Comment body (markdown)"),
				mcplib.Required(),
			),
		),
		h.addIssueComment,
	)

	s.AddTool(
		mcplib.NewTool("edit_issue_comment",
			mcplib.WithDescription("Edit a comment on an issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("New comment body (markdown)"),
				mcplib.Required(),
			),
		),
		h.editIssueComment,
	)

	s.AddTool(
		mcplib.NewTool("delete_issue_comment",
			mcplib.WithDescription("Delete a comment from an issue (destructive; Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteIssueComment,
	)
}
