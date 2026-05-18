package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspaces(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	wc, err := backend.AsWorkspaceClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ws, err := wc.ListWorkspaces(limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(ws)
}

func (h *handlers) listProjects(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	wc, err := backend.AsWorkspaceClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	projects, err := wc.ListProjects(workspace, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(projects)
}
