package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceProjectPermsTools)
}

func registerWorkspaceProjectPermsTools(s *mcpserver.MCPServer, h *handlers) {
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
		mcplib.NewTool("list_workspace_project_perms",
			mcplib.WithDescription("List user and group permissions for a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
		),
		h.listWorkspaceProjectPerms,
	)

	s.AddTool(
		mcplib.NewTool("grant_workspace_project_perm",
			mcplib.WithDescription("Grant a user or group a permission on a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
			mcplib.WithString("permission",
				mcplib.Description("Permission level: read, write, admin, or create-repo"),
				mcplib.Required(),
			),
			mcplib.WithString("user_slug",
				mcplib.Description("User slug to grant permission to (set user_slug or group_slug)"),
			),
			mcplib.WithString("group_slug",
				mcplib.Description("Group slug to grant permission to (set user_slug or group_slug)"),
			),
		),
		h.grantWorkspaceProjectPerm,
	)

	s.AddTool(
		mcplib.NewTool("revoke_workspace_project_perm",
			mcplib.WithDescription("Revoke a user or group permission on a Cloud workspace project"),
			optHostname,
			reqWorkspace,
			reqProjectKey,
			mcplib.WithString("subject_slug",
				mcplib.Description("User or group slug to revoke permission from"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("is_group",
				mcplib.Description("Set to true if subject_slug is a group (default: user)"),
			),
		),
		h.revokeWorkspaceProjectPerm,
	)
}
