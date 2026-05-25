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

	s.AddTool(
		mcplib.NewTool("list_admin_users",
			mcplib.WithDescription("List users on a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
			mcplib.WithString("filter",
				mcplib.Description("Filter users by name or email prefix"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of users to return (default 50, max 1000)"),
			),
		),
		h.listAdminUsers,
	)

	s.AddTool(
		mcplib.NewTool("activate_user",
			mcplib.WithDescription("Activate a user account on Bitbucket Server / DC (Server/DC only)"),
			optHostname,
			mcplib.WithString("slug",
				mcplib.Required(),
				mcplib.Description("Username (slug) of the user to activate"),
			),
		),
		h.activateUser,
	)

	s.AddTool(
		mcplib.NewTool("deactivate_user",
			mcplib.WithDescription("Deactivate a user account on Bitbucket Server / DC (Server/DC only)"),
			optHostname,
			mcplib.WithString("slug",
				mcplib.Required(),
				mcplib.Description("Username (slug) of the user to deactivate"),
			),
		),
		h.deactivateUser,
	)

	s.AddTool(
		mcplib.NewTool("rename_user",
			mcplib.WithDescription("Rename a user (change username/slug) on Bitbucket Server / DC (Server/DC only)"),
			optHostname,
			mcplib.WithString("slug",
				mcplib.Required(),
				mcplib.Description("Current username (slug) of the user"),
			),
			mcplib.WithString("new_slug",
				mcplib.Required(),
				mcplib.Description("New username (slug) for the user"),
			),
		),
		h.renameUser,
	)

	s.AddTool(
		mcplib.NewTool("get_admin_license",
			mcplib.WithDescription("Get license details for a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
		),
		h.getAdminLicense,
	)

	s.AddTool(
		mcplib.NewTool("get_cluster_nodes",
			mcplib.WithDescription("Get cluster node information for a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
		),
		h.getClusterNodes,
	)

	s.AddTool(
		mcplib.NewTool("get_mail_server_config",
			mcplib.WithDescription("Get the mail server configuration for a Bitbucket Server / DC instance (Server/DC only). The password field is never returned."),
			optHostname,
		),
		h.getMailServerConfig,
	)

	s.AddTool(
		mcplib.NewTool("set_mail_server_config",
			mcplib.WithDescription("Update the mail server configuration for a Bitbucket Server / DC instance (Server/DC only)"),
			optHostname,
			mcplib.WithString("mail_hostname",
				mcplib.Required(),
				mcplib.Description("Mail server hostname"),
			),
			mcplib.WithNumber("port",
				mcplib.Description("Mail server port (default 25)"),
			),
			mcplib.WithString("protocol",
				mcplib.Description("Protocol: smtp or smtps (default smtp)"),
			),
			mcplib.WithBoolean("use_starttls",
				mcplib.Description("Enable STARTTLS if available"),
			),
			mcplib.WithBoolean("require_starttls",
				mcplib.Description("Require STARTTLS (fail if not available)"),
			),
			mcplib.WithString("username",
				mcplib.Description("SMTP authentication username"),
			),
			mcplib.WithString("sender_address",
				mcplib.Description("Sender email address (From:)"),
			),
			mcplib.WithString("password",
				mcplib.Description("SMTP password (note: passed in plaintext via MCP)"),
			),
		),
		h.setMailServerConfig,
	)
}
