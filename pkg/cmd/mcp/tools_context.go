package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	contextcmd "github.com/proggarapsody/bitbottle/pkg/cmd/context"
)

func init() {
	registerTool(registerContextTool)
}

func registerContextTool(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("get_context",
			mcplib.WithDescription("Return host, project, slug, branch, default branch, ahead/behind, authenticated user, and backend type in one call. Outside a git repo project/slug/branch/default_branch/ahead/behind are empty."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.getContext,
	)
}

// getContext is the MCP analogue of the `bitbottle context` CLI command. It
// returns one structured snapshot — host, project, slug, branch, default
// branch, ahead/behind counts, authenticated user, backend type — so AI
// agents can orient themselves at the start of a flow with a single tool
// call instead of three (list_hosts / get_repo / get_current_user + git).
func (h *handlers) getContext(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	ctx, err := contextcmd.Build(h.f, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(ctx)
}
