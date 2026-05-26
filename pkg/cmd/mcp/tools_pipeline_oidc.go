package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineOIDCTools)
}

func registerPipelineOIDCTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug (Cloud only)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_oidc_config",
			mcplib.WithDescription("Get the pipeline OIDC discovery document for a workspace"),
			optHostname,
			reqWorkspace,
		),
		h.getPipelineOIDCConfig,
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_oidc_keys",
			mcplib.WithDescription("Get the pipeline OIDC JWKS key set for a workspace"),
			optHostname,
			reqWorkspace,
		),
		h.getPipelineOIDCKeys,
	)
}
