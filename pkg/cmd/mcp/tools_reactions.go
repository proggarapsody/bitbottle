package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerReactionTools)
}

func registerReactionTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_comment_reactions",
			mcplib.WithDescription("List emoji reactions on a pull-request comment, grouped by emoji. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
				mcplib.Required(),
			),
		),
		h.listCommentReactions,
	)

	s.AddTool(
		mcplib.NewTool("add_comment_reaction",
			mcplib.WithDescription("Add an emoji reaction to a pull-request comment. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
				mcplib.Required(),
			),
			mcplib.WithString("emoji",
				mcplib.Description("Emoji shortcode: thumbs_up, thumbs_down, heart, laugh, hooray, confused"),
				mcplib.Required(),
			),
		),
		h.addCommentReaction,
	)

	s.AddTool(
		mcplib.NewTool("remove_comment_reaction",
			mcplib.WithDescription("Remove an emoji reaction from a pull-request comment. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
				mcplib.Required(),
			),
			mcplib.WithString("emoji",
				mcplib.Description("Emoji shortcode: thumbs_up, thumbs_down, heart, laugh, hooray, confused"),
				mcplib.Required(),
			),
		),
		h.removeCommentReaction,
	)
}
