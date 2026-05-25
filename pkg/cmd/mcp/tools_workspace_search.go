package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceSearchTools)
}

func registerWorkspaceSearchTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("search_workspaces",
			mcplib.WithDescription("Search Bitbucket Cloud workspaces by slug/name prefix with optional role filter"),
			mcplib.WithString("query",
				mcplib.Description("Slug/name prefix to match"),
			),
			mcplib.WithString("role",
				mcplib.Description("Filter by role: owner, collaborator, or member"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of workspaces (default 30, 0 = no cap)"),
			),
		),
		h.searchWorkspaces,
	)
}
