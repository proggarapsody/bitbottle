package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspaceMembers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 0)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wmc, err := backend.AsWorkspaceMemberClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	members, err := wmc.ListWorkspaceMembers(workspace, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(members)
}
