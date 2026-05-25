package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineConfigTools)
}

func registerPipelineConfigTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Workspace or project key"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_config",
			mcplib.WithDescription("Get pipeline configuration for a repository"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.getPipelineConfig,
	)

	s.AddTool(
		mcplib.NewTool("enable_pipelines",
			mcplib.WithDescription("Enable pipelines for a repository"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.enablePipelines,
	)

	s.AddTool(
		mcplib.NewTool("disable_pipelines",
			mcplib.WithDescription("Disable pipelines for a repository"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.disablePipelines,
	)
}
