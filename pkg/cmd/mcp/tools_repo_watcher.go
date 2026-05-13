package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoWatcherTools)
}

func registerRepoWatcherTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_repo_watchers",
			mcplib.WithDescription("List users watching a repository"),
			optHostname,
			reqRepo,
		),
		h.listRepoWatchers,
	)
}
