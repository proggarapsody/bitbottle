package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWebhookTools)
}

func registerWebhookTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key or workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_webhooks",
			mcplib.WithDescription("List repository webhooks"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listWebhooks,
	)

	s.AddTool(
		mcplib.NewTool("get_webhook",
			mcplib.WithDescription("Get a single webhook by ID"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("id",
				mcplib.Description("Webhook ID"),
				mcplib.Required(),
			),
		),
		h.getWebhook,
	)

	s.AddTool(
		mcplib.NewTool("create_webhook",
			mcplib.WithDescription("Create a repository webhook"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("url",
				mcplib.Description("Webhook delivery URL"),
				mcplib.Required(),
			),
			mcplib.WithArray("events",
				mcplib.Description("Event keys the webhook subscribes to"),
				mcplib.WithStringItems(),
				mcplib.Required(),
			),
			mcplib.WithBoolean("active",
				mcplib.Description("Whether the webhook is active on creation (defaults to true)"),
			),
			mcplib.WithString("secret",
				mcplib.Description("Shared secret for HMAC signing of delivery payloads (optional)"),
			),
		),
		h.createWebhook,
	)

	s.AddTool(
		mcplib.NewTool("delete_webhook",
			mcplib.WithDescription("Delete a webhook by ID (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("id",
				mcplib.Description("Webhook ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteWebhook,
	)
}
