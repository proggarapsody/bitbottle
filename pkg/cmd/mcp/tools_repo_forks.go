package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoForkTools)
}

func registerRepoForkTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of forks to return (default 30, 0 = no limit)"),
	)

	s.AddTool(
		mcplib.NewTool("list_repo_forks",
			mcplib.WithDescription("List forks of a repository"),
			optHostname,
			reqRepo,
			optLimit,
		),
		h.listRepoForks,
	)
}
