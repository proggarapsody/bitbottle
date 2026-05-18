package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerHostTools)
}

func registerHostTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_hosts",
			mcplib.WithDescription("List all configured Bitbucket hosts"),
		),
		h.listHosts,
	)
}
