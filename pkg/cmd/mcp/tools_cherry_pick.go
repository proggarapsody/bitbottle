package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCherryPickTools)
}

func registerCherryPickTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("cherry_pick_commit",
			mcplib.WithDescription("Cherry-pick a commit onto a target branch (Bitbucket Server / Data Center only)"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository as PROJECT/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("commit_hash",
				mcplib.Description("Hash of the commit to cherry-pick"),
				mcplib.Required(),
			),
			mcplib.WithString("target_branch",
				mcplib.Description("Name of the branch to cherry-pick onto"),
				mcplib.Required(),
			),
			mcplib.WithString("message",
				mcplib.Description("Optional commit message override"),
			),
		),
		h.cherryPickCommit,
	)
}
