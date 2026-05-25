package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerIPAllowlistTools)
}

func registerIPAllowlistTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_workspace_ipallowlists",
			mcplib.WithDescription("List workspace IP allowlist entries (Cloud only)"),
			mcplib.WithString("workspace", mcplib.Description("Workspace slug"), mcplib.Required()),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
		),
		h.listWorkspaceIPAllowlists,
	)

	s.AddTool(
		mcplib.NewTool("create_workspace_ipallowlist",
			mcplib.WithDescription("Add an IP allowlist entry to a workspace (Cloud only)"),
			mcplib.WithString("workspace", mcplib.Description("Workspace slug"), mcplib.Required()),
			mcplib.WithString("cidr", mcplib.Description("IP range in CIDR notation (e.g. 10.0.0.0/8)"), mcplib.Required()),
			mcplib.WithString("description", mcplib.Description("Description for this entry")),
			mcplib.WithString("enabled", mcplib.Description("Whether the entry is enabled (true/false, default true)")),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
		),
		h.createWorkspaceIPAllowlist,
	)

	s.AddTool(
		mcplib.NewTool("delete_workspace_ipallowlist",
			mcplib.WithDescription("Delete an IP allowlist entry from a workspace (Cloud only)"),
			mcplib.WithString("workspace", mcplib.Description("Workspace slug"), mcplib.Required()),
			mcplib.WithString("uuid", mcplib.Description("UUID of the IP allowlist entry to delete"), mcplib.Required()),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
		),
		h.deleteWorkspaceIPAllowlist,
	)
}
