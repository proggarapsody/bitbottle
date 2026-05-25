package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerWorkspacePipelineVarTools)
}

func registerWorkspacePipelineVarTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqKey := mcplib.WithString("key",
		mcplib.Description("Variable key"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_workspace_pipeline_vars",
			mcplib.WithDescription("List workspace-level pipeline variables (Cloud only)"),
			optHostname,
			reqWorkspace,
		),
		h.listWorkspacePipelineVars,
	)

	s.AddTool(
		mcplib.NewTool("get_workspace_pipeline_var",
			mcplib.WithDescription("Get a workspace pipeline variable by key (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
		),
		h.getWorkspacePipelineVar,
	)

	s.AddTool(
		mcplib.NewTool("set_workspace_pipeline_var",
			mcplib.WithDescription("Create or update a workspace pipeline variable (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
			mcplib.WithString("value",
				mcplib.Description("Variable value"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("secured",
				mcplib.Description("Mark as secured (value redacted on read)"),
			),
		),
		h.setWorkspacePipelineVar,
	)

	s.AddTool(
		mcplib.NewTool("delete_workspace_pipeline_var",
			mcplib.WithDescription("Delete a workspace pipeline variable by key (Cloud only)"),
			optHostname,
			reqWorkspace,
			reqKey,
		),
		h.deleteWorkspacePipelineVar,
	)
}
