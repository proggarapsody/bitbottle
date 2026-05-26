package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceProjectDefaultReviewerTools)
}

func registerWorkspaceProjectDefaultReviewerTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqProjectKey := mcplib.WithString("project_key",
		mcplib.Description("Project key (e.g. PROJ)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_project_default_reviewers",
			mcplib.WithDescription("List default reviewers for a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of results (0 = no cap)"),
			),
		),
		h.listProjectDefaultReviewers,
	)

	s.AddTool(
		mcplib.NewTool("add_project_default_reviewer",
			mcplib.WithDescription("Add a default reviewer to a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
			mcplib.WithString("user",
				mcplib.Description("Account ID of the user to add as default reviewer"),
				mcplib.Required(),
			),
		),
		h.addProjectDefaultReviewer,
	)

	s.AddTool(
		mcplib.NewTool("remove_project_default_reviewer",
			mcplib.WithDescription("Remove a default reviewer from a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
			mcplib.WithString("user",
				mcplib.Description("Account ID of the user to remove from default reviewers"),
				mcplib.Required(),
			),
		),
		h.removeProjectDefaultReviewer,
	)
}
