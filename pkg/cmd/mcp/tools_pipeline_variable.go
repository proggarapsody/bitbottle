package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineVariableTools)
}

func registerPipelineVariableTools(s *mcpserver.MCPServer, h *handlers) {
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
	reqKey := mcplib.WithString("key",
		mcplib.Description("Pipeline variable key"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_variables",
			mcplib.WithDescription("List repository-level pipeline variables (Bitbucket Cloud only). Secured variable values are not returned."),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listPipelineVariables,
	)

	s.AddTool(
		mcplib.NewTool("set_pipeline_variable",
			mcplib.WithDescription("Create or update a repository-level pipeline variable, upsert by key (destructive write; Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqKey,
			mcplib.WithString("value",
				mcplib.Description("Variable value"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("secured",
				mcplib.Description("Mark as secured (value redacted on read)"),
			),
		),
		h.setPipelineVariable,
	)

	s.AddTool(
		mcplib.NewTool("delete_pipeline_variable",
			mcplib.WithDescription("Delete a repository-level pipeline variable by key (destructive; Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqKey,
		),
		h.deletePipelineVariable,
	)
}
