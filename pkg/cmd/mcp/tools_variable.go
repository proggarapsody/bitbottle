package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerVariableTools)
}

func registerVariableTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)
	optScope := mcplib.WithString("scope",
		mcplib.Description("Variable scope: repository (default), workspace, or deployment"),
	)
	optEnvUUID := mcplib.WithString("env_uuid",
		mcplib.Description("Environment UUID (required when scope=deployment)"),
	)

	s.AddTool(
		mcplib.NewTool("variable_list",
			mcplib.WithDescription("List variables for a repository, workspace, or deployment environment (Cloud only)"),
			optHostname,
			reqRepo,
			optScope,
			optEnvUUID,
		),
		h.variableList,
	)

	s.AddTool(
		mcplib.NewTool("variable_view",
			mcplib.WithDescription("View a single pipeline variable by key (Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithString("key",
				mcplib.Description("Variable key to look up"),
				mcplib.Required(),
			),
			optScope,
			optEnvUUID,
		),
		h.variableView,
	)

	s.AddTool(
		mcplib.NewTool("variable_set",
			mcplib.WithDescription("Create or update a variable by key (upsert; Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithString("key",
				mcplib.Description("Variable key"),
				mcplib.Required(),
			),
			mcplib.WithString("value",
				mcplib.Description("Variable value"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("secured",
				mcplib.Description("Mark as secured (value redacted on read)"),
			),
			optScope,
			optEnvUUID,
		),
		h.variableSet,
	)

	s.AddTool(
		mcplib.NewTool("variable_delete",
			mcplib.WithDescription("Delete a variable by key (destructive; Cloud only)"),
			optHostname,
			reqRepo,
			mcplib.WithString("key",
				mcplib.Description("Variable key to delete"),
				mcplib.Required(),
			),
			optScope,
			optEnvUUID,
		),
		h.variableDelete,
	)
}
