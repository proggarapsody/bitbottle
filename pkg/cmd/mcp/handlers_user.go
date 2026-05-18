package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) getCurrentUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := client.GetCurrentUser()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(user)
}
