package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRCommitTools)
}

func registerPRCommitTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)
	reqPRID := mcplib.WithNumber("pr_id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pr_commits",
			mcplib.WithDescription("List commits in a pull request"),
			optHostname,
			reqRepo,
			reqPRID,
		),
		h.listPRCommits,
	)
}
