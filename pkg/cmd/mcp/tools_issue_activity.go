package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIssueActivityTools)
}

func registerIssueActivityTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_issue_activity",
			mcplib.WithDescription("List the activity/change history of a Cloud issue"),
			mcplib.WithNumber("issue_id",
				mcplib.Description("Issue ID"),
				mcplib.Required(),
			),
			mcplib.WithString("workspace",
				mcplib.Description("Bitbucket Cloud workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo_slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of activity entries (default 25)"),
			),
		),
		h.listIssueActivity,
	)
}
