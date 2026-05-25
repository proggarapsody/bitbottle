package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceProjectTools)
}

func registerWorkspaceProjectTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqKey := mcplib.WithString("key",
		mcplib.Description("Project key"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("create_workspace_project",
			mcplib.WithDescription("Create a project in a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
			mcplib.WithString("name",
				mcplib.Description("Project name"),
				mcplib.Required(),
			),
			mcplib.WithString("description",
				mcplib.Description("Project description"),
			),
			mcplib.WithBoolean("private",
				mcplib.Description("Make project private"),
			),
		),
		h.createWorkspaceProject,
	)

	s.AddTool(
		mcplib.NewTool("view_workspace_project",
			mcplib.WithDescription("View a workspace project by key (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
		),
		h.viewWorkspaceProject,
	)

	s.AddTool(
		mcplib.NewTool("edit_workspace_project",
			mcplib.WithDescription("Edit a workspace project (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
			mcplib.WithString("name",
				mcplib.Description("New project name"),
			),
			mcplib.WithString("description",
				mcplib.Description("New project description"),
			),
			mcplib.WithBoolean("private",
				mcplib.Description("Set project private flag"),
			),
		),
		h.editWorkspaceProject,
	)

	s.AddTool(
		mcplib.NewTool("delete_workspace_project",
			mcplib.WithDescription("Delete a workspace project (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
		),
		h.deleteWorkspaceProject,
	)
}
