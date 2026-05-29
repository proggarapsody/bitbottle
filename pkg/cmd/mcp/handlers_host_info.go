package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// getHostInfo is the MCP handler for the `get_host_info` tool. Supported by
// both Cloud and Server/DC backends.
func (h *handlers) getHostInfo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	hic, err := backend.AsHostInfoClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	info, err := hic.GetHostInfo()
	if err != nil {
		return errResultErr(err), nil
	}

	return jsonResult(info)
}
