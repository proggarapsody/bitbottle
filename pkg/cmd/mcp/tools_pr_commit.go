package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRCommitTools)
}

func registerPRCommitTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	optProject := mcplib.WithString("project",
		mcplib.Description("Project key (Server) or workspace slug (Cloud)"),
	)
	optSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
	)
	optLegacyRepo := mcplib.WithString("repo",
		mcplib.Description("DEPRECATED: use project + slug. Repository as PROJECT/REPO or WORKSPACE/REPO."),
	)
	reqPRID := mcplib.WithNumber("pr_id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
	)

	addGatedTool(s, h,
		mcplib.NewTool("list_pr_commits",
			mcplib.WithDescription("List commits in a pull request"),
			optHostname,
			optProject,
			optSlug,
			optLegacyRepo,
			reqPRID,
		),
		backendsBoth,
		h.listPRCommits,
	)
}
