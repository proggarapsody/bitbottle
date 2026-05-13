package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineTriggerTools)
}

func registerPipelineTriggerTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("trigger_pipeline",
			mcplib.WithDescription("Trigger a Bitbucket Cloud pipeline on a branch"),
			optHostname,
			reqRepo,
			mcplib.WithString("branch",
				mcplib.Description("Branch to trigger the pipeline on"),
				mcplib.Required(),
			),
			mcplib.WithString("variables",
				mcplib.Description("Pipeline variables as a comma-separated list of key=value pairs (e.g. FOO=bar,BAZ=qux)"),
			),
		),
		h.triggerPipeline,
	)
}
