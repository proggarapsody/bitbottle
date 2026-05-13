package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerDeploymentTools)
}

func registerDeploymentTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)
	reqEnvUUID := mcplib.WithString("env_uuid",
		mcplib.Description("Environment UUID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_deployments",
			mcplib.WithDescription("List deployments for a repository (Bitbucket Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of deployments to return (default 10)"),
			),
		),
		h.listDeployments,
	)

	s.AddTool(
		mcplib.NewTool("get_deployment",
			mcplib.WithDescription("Get a single deployment by UUID (Bitbucket Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithString("uuid",
				mcplib.Description("Deployment UUID"),
				mcplib.Required(),
			),
		),
		h.getDeployment,
	)

	s.AddTool(
		mcplib.NewTool("list_environments",
			mcplib.WithDescription("List deployment environments for a repository (Bitbucket Cloud only)"),
			optHostname,
			reqRepo,
		),
		h.listEnvironments,
	)

	s.AddTool(
		mcplib.NewTool("create_environment",
			mcplib.WithDescription("Create a deployment environment (Bitbucket Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithString("name",
				mcplib.Description("Environment name"),
				mcplib.Required(),
			),
			mcplib.WithString("type",
				mcplib.Description("Environment type: Test, Staging, or Production"),
				mcplib.Required(),
			),
			mcplib.WithNumber("rank",
				mcplib.Description("Numeric ordering rank"),
			),
		),
		h.createEnvironment,
	)

	s.AddTool(
		mcplib.NewTool("delete_environment",
			mcplib.WithDescription("Delete a deployment environment by UUID (destructive; Bitbucket Cloud only)"),
			optHostname,
			reqRepo,
			reqEnvUUID,
		),
		h.deleteEnvironment,
	)

}
