package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerSSHKeyTools)
}

func registerSSHKeyTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)

	s.AddTool(
		mcplib.NewTool("list_ssh_keys",
			mcplib.WithDescription("List SSH keys for the current user"),
			optHostname,
		),
		h.listSSHKeys,
	)

	s.AddTool(
		mcplib.NewTool("add_ssh_key",
			mcplib.WithDescription("Add an SSH key for the current user"),
			optHostname,
			mcplib.WithString("key",
				mcplib.Description("SSH public key string"),
				mcplib.Required(),
			),
			mcplib.WithString("label",
				mcplib.Description("Label for the SSH key"),
			),
		),
		h.addSSHKey,
	)

	s.AddTool(
		mcplib.NewTool("delete_ssh_key",
			mcplib.WithDescription("Delete an SSH key for the current user"),
			optHostname,
			mcplib.WithNumber("id",
				mcplib.Description("SSH key ID"),
				mcplib.Required(),
			),
		),
		h.deleteSSHKey,
	)
}
