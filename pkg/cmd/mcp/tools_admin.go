package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerAdminTools)
}

// registerAdminTools wires the admin tools onto the MCP server.
// All tools are Bitbucket Server / DC only — calls against Cloud return
// host.unsupported via AsAdminClient.
func registerAdminTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)

	s.AddTool(
		mcplib.NewTool("rotate_secrets",
			mcplib.WithDescription("Rotate the cluster HTTPS secret (Bitbucket Server / DC only). All nodes must be restarted after rotation."),
			optHostname,
		),
		h.rotateSecrets,
	)

	s.AddTool(
		mcplib.NewTool("get_logging_config",
			mcplib.WithDescription("Get the current log level and async-logging setting (Bitbucket Server / DC only)"),
			optHostname,
		),
		h.getLoggingConfig,
	)

	s.AddTool(
		mcplib.NewTool("set_logging_config",
			mcplib.WithDescription("Set log level or async logging mode (Bitbucket Server / DC only)"),
			optHostname,
			mcplib.WithString("level",
				mcplib.Description("Log level: DEBUG, INFO, WARN, or ERROR (case-sensitive)"),
			),
			mcplib.WithBoolean("async",
				mcplib.Description("Enable or disable async logging"),
			),
			mcplib.WithBoolean("persistent",
				mcplib.Description("Write to log4j.properties so the change survives restarts"),
			),
		),
		h.setLoggingConfig,
	)
}
