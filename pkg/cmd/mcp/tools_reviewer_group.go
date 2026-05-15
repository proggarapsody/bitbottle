package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerReviewerGroupTools)
}

func registerReviewerGroupTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO (Bitbucket Server / Data Center only)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("pr_reviewer_group_list",
			mcplib.WithDescription("List reviewer group conditions for a repository"),
			optHostname,
			reqRepo,
		),
		h.listReviewerGroups,
	)

	s.AddTool(
		mcplib.NewTool("pr_reviewer_group_add",
			mcplib.WithDescription("Create a reviewer group condition for a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("name",
				mcplib.Description("Name for the reviewer group"),
				mcplib.Required(),
			),
			mcplib.WithString("users",
				mcplib.Description("Comma-separated list of user slugs"),
				mcplib.Required(),
			),
			mcplib.WithNumber("required_approvals",
				mcplib.Description("Required number of approvals (default 1)"),
			),
		),
		h.addReviewerGroup,
	)

	s.AddTool(
		mcplib.NewTool("pr_reviewer_group_remove",
			mcplib.WithDescription("Remove a reviewer group condition from a repository"),
			optHostname,
			reqRepo,
			mcplib.WithNumber("id",
				mcplib.Description("Condition ID to remove"),
				mcplib.Required(),
			),
		),
		h.removeReviewerGroup,
	)
}
