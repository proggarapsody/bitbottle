package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerUserTools)
}

func registerUserTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)

	s.AddTool(
		mcplib.NewTool("get_current_user",
			mcplib.WithDescription("Get the currently authenticated user"),
			optHostname,
		),
		h.getCurrentUser,
	)
}
