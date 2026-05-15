package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineCacheTools)
}

func registerPipelineCacheTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_caches",
			mcplib.WithDescription("List pipeline caches for a repository"),
			optHostname,
			reqRepo,
		),
		h.listPipelineCaches,
	)

	s.AddTool(
		mcplib.NewTool("delete_pipeline_cache",
			mcplib.WithDescription("Delete a pipeline cache from a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("uuid",
				mcplib.Description("Cache UUID"),
				mcplib.Required(),
			),
		),
		h.deletePipelineCache,
	)
}
