package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoSyncTools)
}

func registerRepoSyncTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqNS := mcplib.WithString("ns",
		mcplib.Description("Workspace or project slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	optBranch := mcplib.WithString("branch",
		mcplib.Description("Branch to sync (omit to use the repository's default branch)"),
	)

	s.AddTool(
		mcplib.NewTool("sync_repo",
			mcplib.WithDescription("Sync a Cloud fork branch with its upstream repository"),
			optHostname,
			reqNS,
			reqSlug,
			optBranch,
		),
		h.syncRepo,
	)
}
