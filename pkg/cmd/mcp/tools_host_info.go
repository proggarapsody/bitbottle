package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerHostInfoTools)
}

func registerHostInfoTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)

	s.AddTool(
		mcplib.NewTool("get_host_info",
			mcplib.WithDescription("Return backend type, base URL, version (Server/DC only), and the list of supported feature capabilities for the given host"),
			optHostname,
		),
		h.getHostInfo,
	)
}
