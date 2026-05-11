package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCommitCommentTools)
}

func registerCommitCommentTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_commit_comments",
			mcplib.WithDescription("List all comments on a commit"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of comments to return (0 = no limit)"),
			),
		),
		h.listCommitComments,
	)

	s.AddTool(
		mcplib.NewTool("add_commit_comment",
			mcplib.WithDescription("Add a comment to a commit"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("Comment body"),
				mcplib.Required(),
			),
		),
		h.addCommitComment,
	)

	s.AddTool(
		mcplib.NewTool("edit_commit_comment",
			mcplib.WithDescription("Edit an existing commit comment"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID to edit"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("New comment body"),
				mcplib.Required(),
			),
		),
		h.editCommitComment,
	)

	s.AddTool(
		mcplib.NewTool("delete_commit_comment",
			mcplib.WithDescription("Delete an existing commit comment"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteCommitComment,
	)
}
