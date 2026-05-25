package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoLabelTools)
}

func registerRepoLabelTools(s *mcpserver.MCPServer, h *handlers) {
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
		mcplib.NewTool("list_repo_labels",
			mcplib.WithDescription("List labels on a repository"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listRepoLabels,
	)

	s.AddTool(
		mcplib.NewTool("create_repo_label",
			mcplib.WithDescription("Create a label on a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Label name"),
				mcplib.Required(),
			),
			mcplib.WithString("color",
				mcplib.Description("Label color (hex, e.g. #ff0000)"),
			),
		),
		h.createRepoLabel,
	)

	s.AddTool(
		mcplib.NewTool("update_repo_label",
			mcplib.WithDescription("Update a label on a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Label ID"),
				mcplib.Required(),
			),
			mcplib.WithString("name",
				mcplib.Description("New label name"),
			),
			mcplib.WithString("color",
				mcplib.Description("New label color (hex)"),
			),
		),
		h.updateRepoLabel,
	)

	s.AddTool(
		mcplib.NewTool("delete_repo_label",
			mcplib.WithDescription("Delete a label from a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Label ID"),
				mcplib.Required(),
			),
		),
		h.deleteRepoLabel,
	)
}
