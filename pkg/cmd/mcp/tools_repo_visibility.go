package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoVisibilityTools)
}

func registerRepoVisibilityTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("repo_visibility",
			mcplib.WithDescription("Get or set repository visibility (public or private)"),
			mcplib.WithString("repo",
				mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("visibility",
				mcplib.Description("Omit to get current visibility; provide \"public\" or \"private\" to set"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.repoVisibility,
	)
}
