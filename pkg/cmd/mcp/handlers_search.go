package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// searchCode is the MCP handler for the `search_code` tool. Cloud-only —
// the AsCodeSearcher type-assertion stamps the host.unsupported envelope
// when the backend is Server / DC.
func (h *handlers) searchCode(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	query, err := requireString(req, "query")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cs, err := backend.AsCodeSearcher(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hits, err := cs.SearchCode(workspace, query, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hits)
}
