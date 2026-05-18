package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerTagTools)
}

func registerTagTools(s *mcpserver.MCPServer, h *handlers) {
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
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results to return"),
	)

	s.AddTool(
		mcplib.NewTool("list_tags",
			mcplib.WithDescription("List tags for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			optLimit,
		),
		h.listTags,
	)

	s.AddTool(
		mcplib.NewTool("create_tag",
			mcplib.WithDescription("Create a tag in a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Tag name"),
				mcplib.Required(),
			),
			mcplib.WithString("start_at",
				mcplib.Description("Branch name or commit hash to tag"),
				mcplib.Required(),
			),
			mcplib.WithString("message",
				mcplib.Description("Tag message (creates annotated tag when non-empty)"),
			),
		),
		h.createTag,
	)

	s.AddTool(
		mcplib.NewTool("delete_tag",
			mcplib.WithDescription("Delete a tag (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Tag name to delete"),
				mcplib.Required(),
			),
		),
		h.deleteTag,
	)
}
