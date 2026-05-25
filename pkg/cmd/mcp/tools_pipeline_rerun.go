package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineRerunTools)
}

func registerPipelineRerunTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("rerun_pipeline",
			mcplib.WithDescription("Re-run a Bitbucket Cloud pipeline at the same commit as a previous run"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository as WORKSPACE/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("pipeline_uuid",
				mcplib.Description("UUID of the pipeline to rerun"),
				mcplib.Required(),
			),
		),
		h.rerunPipeline,
	)
}
