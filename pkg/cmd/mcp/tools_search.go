package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerSearchTools)
}

func registerSearchTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results to return"),
	)

	s.AddTool(
		mcplib.NewTool("search_code",
			mcplib.WithDescription("Search code across a Bitbucket Cloud workspace (Cloud only). Bitbucket's query language is passed through verbatim — operators like 'path:', 'lang:', and exact-phrase quoting work as documented."),
			optHostname,
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug to scope the search to"),
				mcplib.Required(),
			),
			mcplib.WithString("query",
				mcplib.Description("Bitbucket Cloud code-search query string"),
				mcplib.Required(),
			),
			optLimit,
		),
		h.searchCode,
	)
}
