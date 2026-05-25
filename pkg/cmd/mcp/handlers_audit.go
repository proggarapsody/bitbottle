package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspaceAuditLog(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ac, err := backend.AsAuditClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	opts := backend.AuditLogOpts{
		Action: req.GetString("action", ""),
		From:   req.GetString("from", ""),
		Limit:  req.GetInt("limit", 25),
	}
	events, err := ac.ListAuditLog(workspace, opts)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(events)
}
