package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineStopTools)
}

func registerPipelineStopTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("stop_pipeline",
			mcplib.WithDescription("Stop a running Bitbucket Cloud pipeline"),
			optHostname,
			reqRepo,
			mcplib.WithString("uuid",
				mcplib.Description("Pipeline UUID to stop"),
				mcplib.Required(),
			),
		),
		h.stopPipeline,
	)
}
