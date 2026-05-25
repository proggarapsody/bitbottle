package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIssueVersionTools)
}

func registerIssueVersionTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_issue_versions",
			mcplib.WithDescription("List issue versions for a repository (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSlug,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of versions (default 30)"),
			),
		),
		h.listIssueVersions,
	)

	s.AddTool(
		mcplib.NewTool("view_issue_version",
			mcplib.WithDescription("View a single issue version by ID (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Version ID"),
				mcplib.Required(),
			),
		),
		h.viewIssueVersion,
	)

	s.AddTool(
		mcplib.NewTool("create_issue_version",
			mcplib.WithDescription("Create an issue version in a repository (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Version name"),
				mcplib.Required(),
			),
		),
		h.createIssueVersion,
	)

	s.AddTool(
		mcplib.NewTool("delete_issue_version",
			mcplib.WithDescription("Delete an issue version by ID (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Version ID"),
				mcplib.Required(),
			),
		),
		h.deleteIssueVersion,
	)
}
