package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspaceHookTools)
}

func registerWorkspaceHookTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("workspace_hook_list",
			mcplib.WithDescription("List workspace-level webhooks (Bitbucket Cloud only)"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
		),
		h.listWorkspaceHooks,
	)

	s.AddTool(
		mcplib.NewTool("workspace_hook_create",
			mcplib.WithDescription("Create a workspace-level webhook (Bitbucket Cloud only)"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("url",
				mcplib.Description("Webhook URL"),
				mcplib.Required(),
			),
			mcplib.WithString("events",
				mcplib.Description("Comma-separated list of events to subscribe to"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("active",
				mcplib.Description("Whether the webhook is active (default true)"),
			),
		),
		h.createWorkspaceHook,
	)

	s.AddTool(
		mcplib.NewTool("workspace_hook_delete",
			mcplib.WithDescription("Delete a workspace-level webhook (Bitbucket Cloud only)"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("uuid",
				mcplib.Description("Webhook UUID"),
				mcplib.Required(),
			),
		),
		h.deleteWorkspaceHook,
	)
}
