package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPATTools)
}

func registerPATTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_pats",
			mcplib.WithDescription("List personal access tokens for a user on Bitbucket Server/DC. Returns an error on Cloud."),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of tokens to return (default 50)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.listPATs,
	)
	s.AddTool(
		mcplib.NewTool("create_pat",
			mcplib.WithDescription("Create a personal access token on Bitbucket Server/DC. Returns an error on Cloud. IMPORTANT: The response contains a secret token field that will not be shown again — store it immediately."),
			mcplib.WithString("name",
				mcplib.Required(),
				mcplib.Description("Token name"),
			),
			mcplib.WithString("scopes",
				mcplib.Required(),
				mcplib.Description("Comma-separated permission scopes: repo:read, repo:write, pr:read, pr:write, project:read, project:write"),
			),
			mcplib.WithNumber("expires_in",
				mcplib.Description("Token lifetime in days (omit for no expiry)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.createPAT,
	)
	s.AddTool(
		mcplib.NewTool("revoke_pat",
			mcplib.WithDescription("Revoke a personal access token on Bitbucket Server/DC by token ID. Returns an error on Cloud."),
			mcplib.WithString("token_id",
				mcplib.Required(),
				mcplib.Description("Token ID to revoke"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.revokePAT,
	)
}
