package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRefCompareTools)
}

func registerRefCompareTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("compare_refs",
			mcplib.WithDescription("Compare two branches or commits; returns ahead/behind counts and commit lists"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository in WORKSPACE/REPO or PROJECT/REPO form"),
				mcplib.Required(),
			),
			mcplib.WithString("base",
				mcplib.Description("Base branch or commit"),
				mcplib.Required(),
			),
			mcplib.WithString("head",
				mcplib.Description("Head branch or commit"),
				mcplib.Required(),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of commits to return per side (default 30)"),
			),
		),
		h.compareRefs,
	)
}
