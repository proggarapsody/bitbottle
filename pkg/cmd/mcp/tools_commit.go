package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCommitTools)
}

func registerCommitTools(s *mcpserver.MCPServer, h *handlers) {
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

	s.AddTool(
		mcplib.NewTool("list_commits",
			mcplib.WithDescription("List commits for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("branch",
				mcplib.Description("Branch to list commits from (default: main)"),
			),
			optLimit,
		),
		h.listCommits,
	)

	s.AddTool(
		mcplib.NewTool("get_commit",
			mcplib.WithDescription("Get a single commit by hash"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
		),
		h.getCommit,
	)

	s.AddTool(
		mcplib.NewTool("list_commit_statuses",
			mcplib.WithDescription("List build / CI statuses reported against a commit hash"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
		),
		h.listCommitStatuses,
	)
}
