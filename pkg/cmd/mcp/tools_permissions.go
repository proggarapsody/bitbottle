package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPermissionsTools)
}

// registerPermissionsTools wires the permissions tools onto the MCP server.
// All tools are Bitbucket Server / DC only — calls against Cloud return
// host.unsupported via AsPermissionsClient.
func registerPermissionsTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key"),
		mcplib.Required(),
	)
	optSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	optUser := mcplib.WithString("user",
		mcplib.Description("User slug (specify user or group, not both)"),
	)
	optGroup := mcplib.WithString("group",
		mcplib.Description("Group name (specify user or group, not both)"),
	)
	reqPerm := mcplib.WithString("permission",
		mcplib.Description("Permission level (PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN, REPO_READ, REPO_WRITE, REPO_ADMIN)"),
		mcplib.Required(),
	)

	// ── Project permissions ───────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_project_permissions",
			mcplib.WithDescription("List user and group permissions for a project (Bitbucket Server / DC only)"),
			optHostname, reqProject,
		),
		h.listProjectPermissions,
	)

	s.AddTool(
		mcplib.NewTool("grant_project_permission",
			mcplib.WithDescription("Grant a user or group a permission on a project (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqPerm, optUser, optGroup,
		),
		h.grantProjectPermission,
	)

	s.AddTool(
		mcplib.NewTool("revoke_project_permission",
			mcplib.WithDescription("Revoke a user or group permission on a project (Bitbucket Server / DC only)"),
			optHostname, reqProject, optUser, optGroup,
		),
		h.revokeProjectPermission,
	)

	// ── Repo permissions ──────────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_repo_permissions",
			mcplib.WithDescription("List user and group permissions for a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, optSlug,
		),
		h.listRepoPermissions,
	)

	s.AddTool(
		mcplib.NewTool("grant_repo_permission",
			mcplib.WithDescription("Grant a user or group a permission on a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, optSlug, reqPerm, optUser, optGroup,
		),
		h.grantRepoPermission,
	)

	s.AddTool(
		mcplib.NewTool("revoke_repo_permission",
			mcplib.WithDescription("Revoke a user or group permission on a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, optSlug, optUser, optGroup,
		),
		h.revokeRepoPermission,
	)
}
