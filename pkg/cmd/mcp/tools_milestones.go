package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerMilestoneTools)
}

func registerMilestoneTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_milestones",
			mcplib.WithDescription("List issue milestones for a repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of milestones (default 30)"),
			),
		),
		h.listMilestones,
	)

	s.AddTool(
		mcplib.NewTool("view_milestone",
			mcplib.WithDescription("View a single issue milestone by ID (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Milestone ID"),
				mcplib.Required(),
			),
		),
		h.viewMilestone,
	)
}
