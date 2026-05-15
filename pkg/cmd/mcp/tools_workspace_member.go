package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceMemberTools)
}

func registerWorkspaceMemberTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_workspace_members",
			mcplib.WithDescription("List members of a Bitbucket Cloud workspace"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of members to return (0 = no cap)"),
			),
		),
		h.listWorkspaceMembers,
	)
}
