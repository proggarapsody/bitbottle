package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIssueExtrasTools)
}

func registerIssueExtrasTools(s *mcpserver.MCPServer, h *handlers) {
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
		mcplib.NewTool("list_issue_attachments",
			mcplib.WithDescription("List attachments on a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.listIssueAttachments,
	)

	s.AddTool(
		mcplib.NewTool("delete_issue_attachment",
			mcplib.WithDescription("Delete an attachment from a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithString("filename",
				mcplib.Description("Filename of the attachment to delete"),
				mcplib.Required(),
			),
		),
		h.deleteIssueAttachment,
	)

	s.AddTool(
		mcplib.NewTool("vote_issue",
			mcplib.WithDescription("Vote on a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.voteIssue,
	)

	s.AddTool(
		mcplib.NewTool("unvote_issue",
			mcplib.WithDescription("Remove vote from a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.unvoteIssue,
	)

	s.AddTool(
		mcplib.NewTool("watch_issue",
			mcplib.WithDescription("Watch a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.watchIssue,
	)

	s.AddTool(
		mcplib.NewTool("unwatch_issue",
			mcplib.WithDescription("Stop watching a Cloud issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.unwatchIssue,
	)
}
