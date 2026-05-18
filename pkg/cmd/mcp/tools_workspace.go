package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceTools)
}

func registerWorkspaceTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results to return"),
	)

	s.AddTool(
		mcplib.NewTool("list_workspaces",
			mcplib.WithDescription("List Bitbucket Cloud workspaces the authenticated user belongs to (Cloud only)"),
			optHostname,
			optLimit,
		),
		h.listWorkspaces,
	)

	s.AddTool(
		mcplib.NewTool("list_projects",
			mcplib.WithDescription("List projects within a Bitbucket Cloud workspace (Cloud only)"),
			optHostname,
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
			optLimit,
		),
		h.listProjects,
	)
}
