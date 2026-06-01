package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCommitSearchTools)
}

func registerCommitSearchTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key or workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("search_commits",
			mcplib.WithDescription("Search commits by message keyword, author, or date range"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("query",
				mcplib.Description("Message keyword filter"),
			),
			mcplib.WithString("author",
				mcplib.Description("Author slug or display name"),
			),
			mcplib.WithString("since",
				mcplib.Description("Start date or commit SHA (ISO 8601)"),
			),
			mcplib.WithString("until",
				mcplib.Description("End date or commit SHA (ISO 8601)"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of results to return"),
			),
		),
		h.searchCommits,
	)
}
