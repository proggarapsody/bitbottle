package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) listHosts(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	cfg, err := h.f.Config()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cfg.Hosts())
}
