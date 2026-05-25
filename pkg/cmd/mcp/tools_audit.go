package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerAuditTools)
}

func registerAuditTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_workspace_audit_log",
			mcplib.WithDescription("List workspace audit log events"),
			mcplib.WithString("workspace", mcplib.Description("Workspace slug"), mcplib.Required()),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
			mcplib.WithString("action", mcplib.Description("Filter by action type (e.g. workspace.member.create)")),
			mcplib.WithString("from", mcplib.Description("Return events at or after this ISO 8601 datetime")),
			mcplib.WithNumber("limit", mcplib.Description("Maximum number of events to return (default 25)")),
		),
		h.listWorkspaceAuditLog,
	)
}
