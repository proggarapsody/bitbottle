package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoTransferTools)
}

func registerRepoTransferTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("transfer_repo",
			mcplib.WithDescription("Transfer a repository to another project (Server) or workspace (Cloud)"),
			mcplib.WithString("repo",
				mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("target",
				mcplib.Description("Target project key (Server) or workspace slug (Cloud)"),
				mcplib.Required(),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.transferRepo,
	)
}
