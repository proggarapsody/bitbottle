package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerSnippetCommentTools)
}

func registerSnippetCommentTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqSnippetID := mcplib.WithString("snippet_id",
		mcplib.Description("Snippet ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_snippet_comments",
			mcplib.WithDescription("List comments on a Bitbucket Cloud snippet (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSnippetID,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of results to return"),
			),
		),
		h.listSnippetComments,
	)

	s.AddTool(
		mcplib.NewTool("add_snippet_comment",
			mcplib.WithDescription("Add a comment to a Bitbucket Cloud snippet (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSnippetID,
			mcplib.WithString("body",
				mcplib.Description("Comment body text"),
				mcplib.Required(),
			),
		),
		h.addSnippetComment,
	)

	s.AddTool(
		mcplib.NewTool("delete_snippet_comment",
			mcplib.WithDescription("Delete a comment from a Bitbucket Cloud snippet (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSnippetID,
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteSnippetComment,
	)
}
