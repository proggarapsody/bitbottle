package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerAdminRateLimitTools)
}

// registerAdminRateLimitTools wires the admin rate-limit tools onto the MCP server.
// Both tools are Bitbucket Server / DC only — calls against Cloud return
// host.unsupported via AsAdminClient.
func registerAdminRateLimitTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)

	s.AddTool(
		mcplib.NewTool("get_rate_limit_config",
			mcplib.WithDescription("Get the rate-limiting configuration for a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
		),
		h.getRateLimitConfig,
	)

	s.AddTool(
		mcplib.NewTool("set_rate_limit_config",
			mcplib.WithDescription("Update the rate-limiting configuration for a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
			mcplib.WithBoolean("enabled",
				mcplib.Description("Enable or disable rate limiting"),
			),
			mcplib.WithNumber("requests_per_hour",
				mcplib.Description("Maximum requests per hour per user (omit to keep current value)"),
			),
			mcplib.WithNumber("throttle_wait_ms",
				mcplib.Description("Milliseconds to wait when throttled (omit to keep current value)"),
			),
		),
		h.setRateLimitConfig,
	)
}
