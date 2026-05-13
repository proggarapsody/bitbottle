package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerDiffTools)
}

func registerDiffTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)
	reqFrom := mcplib.WithString("from",
		mcplib.Description("Source ref (branch, tag, or commit hash)"),
		mcplib.Required(),
	)
	reqTo := mcplib.WithString("to",
		mcplib.Description("Target ref (branch, tag, or commit hash)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("get_diff",
			mcplib.WithDescription("Get the unified diff between two refs in a repository"),
			optHostname,
			reqRepo,
			reqFrom,
			reqTo,
		),
		h.getDiff,
	)

	s.AddTool(
		mcplib.NewTool("get_diff_stat",
			mcplib.WithDescription("Get a summary of files changed between two refs in a repository"),
			optHostname,
			reqRepo,
			reqFrom,
			reqTo,
		),
		h.getDiffStat,
	)
}
