package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerDeployKeyTools)
}

func registerDeployKeyTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_deploy_keys",
			mcplib.WithDescription("List deploy keys for a repository"),
			optHostname,
			reqRepo,
		),
		h.listDeployKeys,
	)

	s.AddTool(
		mcplib.NewTool("add_deploy_key",
			mcplib.WithDescription("Add a deploy key to a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("key",
				mcplib.Description("SSH public key string"),
				mcplib.Required(),
			),
			mcplib.WithString("label",
				mcplib.Description("Label for the deploy key"),
			),
		),
		h.addDeployKey,
	)

	s.AddTool(
		mcplib.NewTool("delete_deploy_key",
			mcplib.WithDescription("Delete a deploy key from a repository"),
			optHostname,
			reqRepo,
			mcplib.WithNumber("id",
				mcplib.Description("Deploy key ID"),
				mcplib.Required(),
			),
		),
		h.deleteDeployKey,
	)
}
