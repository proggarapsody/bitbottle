package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerServerProjectTools)
}

func registerServerProjectTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_server_projects",
			mcplib.WithDescription("List projects on a Bitbucket Server/DC instance. Returns an error on Cloud."),
			mcplib.WithString("filter",
				mcplib.Description("Filter projects by name prefix"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of projects to return (default 30)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.listServerProjects,
	)
	s.AddTool(
		mcplib.NewTool("get_server_project",
			mcplib.WithDescription("Get a single project by key on Bitbucket Server/DC. Returns an error on Cloud."),
			mcplib.WithString("key",
				mcplib.Required(),
				mcplib.Description("Project key (e.g. PRJ)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.getServerProject,
	)
	s.AddTool(
		mcplib.NewTool("create_server_project",
			mcplib.WithDescription("Create a new project on Bitbucket Server/DC. Returns an error on Cloud."),
			mcplib.WithString("key",
				mcplib.Required(),
				mcplib.Description("Project key (short uppercase identifier, e.g. PRJ)"),
			),
			mcplib.WithString("name",
				mcplib.Required(),
				mcplib.Description("Project display name"),
			),
			mcplib.WithString("description",
				mcplib.Description("Project description"),
			),
			mcplib.WithBoolean("public",
				mcplib.Description("Make the project publicly accessible (default false)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.createServerProject,
	)
	s.AddTool(
		mcplib.NewTool("update_server_project",
			mcplib.WithDescription("Update a project on Bitbucket Server/DC. Returns an error on Cloud."),
			mcplib.WithString("key",
				mcplib.Required(),
				mcplib.Description("Project key to update"),
			),
			mcplib.WithString("name",
				mcplib.Description("New project display name"),
			),
			mcplib.WithString("description",
				mcplib.Description("New project description"),
			),
			mcplib.WithBoolean("public",
				mcplib.Description("Set project public visibility"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.updateServerProject,
	)
	s.AddTool(
		mcplib.NewTool("delete_server_project",
			mcplib.WithDescription("Delete a project on Bitbucket Server/DC. Returns an error on Cloud."),
			mcplib.WithString("key",
				mcplib.Required(),
				mcplib.Description("Project key to delete"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.deleteServerProject,
	)
}
