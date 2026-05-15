package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerSuggestionTools)
}

func registerSuggestionTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("pr_suggestion_apply",
			mcplib.WithDescription("Apply a Bitbucket Server / DC suggested-change block. The server commits the change to the PR source branch. Use preview=true to show the suggestion body without applying. Bitbucket Cloud is not supported."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID containing the suggestion"),
				mcplib.Required(),
			),
			mcplib.WithNumber("suggestion_id",
				mcplib.Description("Suggestion ID within the comment"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("preview",
				mcplib.Description("Show the suggestion body without applying it"),
			),
		),
		h.prSuggestionApply,
	)
}
