package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerSnippetTools)
}

func registerSnippetTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqSnippetID := mcplib.WithString("id",
		mcplib.Description("Snippet ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_snippets",
			mcplib.WithDescription("List snippets in a Bitbucket Cloud workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of results to return"),
			),
		),
		h.listSnippets,
	)

	s.AddTool(
		mcplib.NewTool("view_snippet",
			mcplib.WithDescription("Get a single snippet by ID (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSnippetID,
		),
		h.viewSnippet,
	)

	s.AddTool(
		mcplib.NewTool("create_snippet",
			mcplib.WithDescription("Create a new snippet in a Bitbucket Cloud workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithString("title",
				mcplib.Description("Snippet title"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("private",
				mcplib.Description("Set true to make the snippet private"),
			),
			mcplib.WithString("files",
				mcplib.Description("JSON object mapping filename to content, e.g. {\"hello.go\": \"package main\"}"),
			),
		),
		h.createSnippet,
	)

	s.AddTool(
		mcplib.NewTool("delete_snippet",
			mcplib.WithDescription("Delete a snippet (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSnippetID,
		),
		h.deleteSnippet,
	)
}
