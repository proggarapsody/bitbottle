package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerBranchModelTools)
}

func registerBranchModelTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("get_branch_model",
			mcplib.WithDescription("Get the effective branching model for a repository"),
			optHostname,
			reqRepo,
		),
		h.getBranchModel,
	)

	s.AddTool(
		mcplib.NewTool("set_branch_model",
			mcplib.WithDescription("Update branching model settings for a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("dev_branch",
				mcplib.Description("Development branch name"),
			),
			mcplib.WithString("prod_branch",
				mcplib.Description("Production branch name"),
			),
			mcplib.WithBoolean("prod_enabled",
				mcplib.Description("Enable the production branch"),
			),
			mcplib.WithObject("branch_type_prefixes",
				mcplib.Description("Map of branch kind to prefix, e.g. {\"feature\": \"feat/\", \"hotfix\": \"hf/\"}"),
			),
		),
		h.setBranchModel,
	)
}
