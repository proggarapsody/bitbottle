package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerMirrorTools)
}

func registerMirrorTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results (0 = no cap)"),
	)

	s.AddTool(
		mcplib.NewTool("list_mirror_servers",
			mcplib.WithDescription("List Smart Mirror servers (Server/DC only)"),
			optHostname,
			optLimit,
		),
		h.listMirrorServers,
	)

	s.AddTool(
		mcplib.NewTool("view_mirror_server",
			mcplib.WithDescription("Get details of a Smart Mirror server (Server/DC only)"),
			optHostname,
			mcplib.WithString("id",
				mcplib.Description("Mirror server ID"),
				mcplib.Required(),
			),
		),
		h.viewMirrorServer,
	)

	s.AddTool(
		mcplib.NewTool("list_mirrored_repos",
			mcplib.WithDescription("List repos mirrored by a Smart Mirror server (Server/DC only)"),
			optHostname,
			mcplib.WithString("mirror_id",
				mcplib.Description("Mirror server ID"),
				mcplib.Required(),
			),
			optLimit,
		),
		h.listMirroredRepos,
	)
}
