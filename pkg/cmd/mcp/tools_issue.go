package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIssueTools)
}

func registerIssueTools(s *mcpserver.MCPServer, h *handlers) {
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
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results to return"),
	)
	reqIssueID := mcplib.WithNumber("id",
		mcplib.Description("Issue ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_issues",
			mcplib.WithDescription("List issues in a Bitbucket Cloud repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("state",
				mcplib.Description("State filter (open, new, on hold, resolved, duplicate, invalid, wontfix, closed); empty = all"),
			),
			optLimit,
		),
		h.listIssues,
	)

	s.AddTool(
		mcplib.NewTool("get_issue",
			mcplib.WithDescription("Get a single issue by ID (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Issue ID"),
				mcplib.Required(),
			),
		),
		h.getIssue,
	)

	s.AddTool(
		mcplib.NewTool("create_issue",
			mcplib.WithDescription("Create a new issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("title",
				mcplib.Description("Issue title"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("Issue body (markdown)"),
			),
			mcplib.WithString("kind",
				mcplib.Description("bug, enhancement, proposal, task"),
			),
			mcplib.WithString("priority",
				mcplib.Description("trivial, minor, major, critical, blocker"),
			),
		),
		h.createIssue,
	)

	s.AddTool(
		mcplib.NewTool("close_issue",
			mcplib.WithDescription("Close an issue by setting its state to \"closed\" (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Issue ID"),
				mcplib.Required(),
			),
		),
		h.closeIssue,
	)

	s.AddTool(
		mcplib.NewTool("update_issue",
			mcplib.WithDescription("Update an issue's title, body, kind, priority, assignee, or state (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithString("title",
				mcplib.Description("New issue title"),
			),
			mcplib.WithString("body",
				mcplib.Description("New issue body (markdown)"),
			),
			mcplib.WithString("kind",
				mcplib.Description("bug, enhancement, proposal, task"),
			),
			mcplib.WithString("priority",
				mcplib.Description("trivial, minor, major, critical, blocker"),
			),
			mcplib.WithString("assignee",
				mcplib.Description("Assignee username; use \"__none__\" to clear the assignee"),
			),
			mcplib.WithString("state",
				mcplib.Description("new, open, resolved, on hold, invalid, duplicate, wontfix, closed"),
			),
		),
		h.updateIssue,
	)

	s.AddTool(
		mcplib.NewTool("reopen_issue",
			mcplib.WithDescription("Reopen a closed issue (sets state to \"open\"; Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
		),
		h.reopenIssue,
	)

	s.AddTool(
		mcplib.NewTool("assign_issue",
			mcplib.WithDescription("Assign an issue to a user by username (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqIssueID,
			mcplib.WithString("assignee",
				mcplib.Description("Bitbucket Cloud username to assign to"),
				mcplib.Required(),
			),
		),
		h.assignIssue,
	)
}
