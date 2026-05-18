package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCommitCommentCoreTools)
}

func registerCommitCommentCoreTools(s *mcpserver.MCPServer, h *handlers) {
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
	reqHash := mcplib.WithString("hash",
		mcplib.Description("Commit hash"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_commit_comments",
			mcplib.WithDescription("List comments on a commit"),
			optHostname,
			reqProject,
			reqSlug,
			reqHash,
			mcplib.WithBoolean("include_reactions",
				mcplib.Description("Fetch and include emoji reactions for each comment (Bitbucket Server / DC only; default: false)"),
			),
		),
		h.listCommitComments,
	)

	s.AddTool(
		mcplib.NewTool("add_commit_comment",
			mcplib.WithDescription("Add a comment to a commit"),
			optHostname,
			reqProject,
			reqSlug,
			reqHash,
			mcplib.WithString("body",
				mcplib.Description("Comment body"),
				mcplib.Required(),
			),
		),
		h.addCommitComment,
	)

	s.AddTool(
		mcplib.NewTool("edit_commit_comment",
			mcplib.WithDescription("Edit a commit comment"),
			optHostname,
			reqProject,
			reqSlug,
			reqHash,
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
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
			mcplib.WithDescription("Delete a commit comment (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			reqHash,
			mcplib.WithNumber("comment_id",
				mcplib.Description("Comment ID"),
				mcplib.Required(),
			),
		),
		h.deleteCommitComment,
	)
}
