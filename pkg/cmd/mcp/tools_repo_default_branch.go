package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoDefaultBranchTools)
}

func registerRepoDefaultBranchTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("set_repo_default_branch",
			mcplib.WithDescription("Set the default branch of a repository"),
			mcplib.WithString("repo",
				mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("branch",
				mcplib.Description("Branch name to set as the default branch"),
				mcplib.Required(),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.handleSetRepoDefaultBranch,
	)
}
