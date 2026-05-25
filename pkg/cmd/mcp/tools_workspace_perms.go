package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspacePermsTools)
}

func registerWorkspacePermsTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results (0 = no cap)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_workspace_perms",
			mcplib.WithDescription("List workspace member permissions (Cloud only)"),
			optHostname,
			reqWorkspace,
			optLimit,
		),
		h.listWorkspacePerms,
	)

	s.AddTool(
		mcplib.NewTool("list_workspace_repo_perms",
			mcplib.WithDescription("List workspace repository permissions (Cloud only)"),
			optHostname,
			reqWorkspace,
			optLimit,
		),
		h.listWorkspaceRepoPerms,
	)

	s.AddTool(
		mcplib.NewTool("grant_workspace_perm",
			mcplib.WithDescription("Grant a user a permission in a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithString("user",
				mcplib.Description("User nickname to grant permission to"),
				mcplib.Required(),
			),
			mcplib.WithString("permission",
				mcplib.Description("Permission level: member, collaborator, or owner"),
				mcplib.Required(),
			),
		),
		h.grantWorkspacePerm,
	)

	s.AddTool(
		mcplib.NewTool("revoke_workspace_perm",
			mcplib.WithDescription("Revoke a user's permission in a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithString("user",
				mcplib.Description("User nickname to revoke permission from"),
				mcplib.Required(),
			),
		),
		h.revokeWorkspacePerm,
	)
}
