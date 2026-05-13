package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerDefaultReviewerTools)
}

func registerDefaultReviewerTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO (Server) or WORKSPACE/REPO (Cloud)"),
		mcplib.Required(),
	)
	reqUser := mcplib.WithString("user",
		mcplib.Description("User slug (Server) or account ID / nickname (Cloud)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_default_reviewers",
			mcplib.WithDescription("List default reviewers for a repository"),
			optHostname,
			reqRepo,
		),
		h.listDefaultReviewers,
	)

	s.AddTool(
		mcplib.NewTool("add_default_reviewer",
			mcplib.WithDescription("Add a default reviewer to a repository"),
			optHostname,
			reqRepo,
			reqUser,
		),
		h.addDefaultReviewer,
	)

	s.AddTool(
		mcplib.NewTool("remove_default_reviewer",
			mcplib.WithDescription("Remove a default reviewer from a repository"),
			optHostname,
			reqRepo,
			reqUser,
		),
		h.removeDefaultReviewer,
	)
}
