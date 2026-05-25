package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRunnerTools)
}

// registerRunnerTools wires the pipeline runner tools onto the MCP server.
// All tools are Bitbucket Cloud only — calls against Server/DC return
// host.unsupported via AsRunnerClient.
func registerRunnerTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("workspace",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_runners",
			mcplib.WithDescription("List Bitbucket Cloud Pipelines self-hosted runners for a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
		),
		h.listRunners,
	)

	s.AddTool(
		mcplib.NewTool("create_runner",
			mcplib.WithDescription("Register a new Bitbucket Cloud Pipelines self-hosted runner in a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithString("name",
				mcplib.Description("Runner name"),
				mcplib.Required(),
			),
			mcplib.WithString("platform",
				mcplib.Description("Runner platform: linux_amd64, linux_arm64, windows_amd64, macos_arm64 (default linux_amd64)"),
			),
			mcplib.WithArray("labels",
				mcplib.Description("Runner labels (array of strings)"),
			),
		),
		h.createRunner,
	)

	s.AddTool(
		mcplib.NewTool("delete_runner",
			mcplib.WithDescription("Remove a Bitbucket Cloud Pipelines self-hosted runner from a workspace (Cloud only)"),
			optHostname,
			reqWorkspace,
			mcplib.WithString("uuid",
				mcplib.Description("Runner UUID"),
				mcplib.Required(),
			),
		),
		h.deleteRunner,
	)
}
