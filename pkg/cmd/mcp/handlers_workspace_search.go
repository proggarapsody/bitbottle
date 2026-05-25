package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) searchWorkspaces(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	query := req.GetString("query", "")
	role := req.GetString("role", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wc, err := backend.AsWorkspaceSearcher(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	workspaces, err := wc.SearchWorkspaces(backend.WorkspaceSearchOpts{
		Query: query,
		Role:  role,
		Limit: limit,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(workspaces)
}
