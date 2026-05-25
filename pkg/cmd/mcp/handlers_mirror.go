package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveMirrorClient is the shared preamble for all mirror handlers:
// parse hostname, dial backend, type-assert MirrorClient.
func (h *handlers) resolveMirrorClient(req mcplib.CallToolRequest) (backend.MirrorClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	mc, err := backend.AsMirrorClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return mc, hostname, nil
}

func (h *handlers) listMirrorServers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	mc, _, err := h.resolveMirrorClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 30)
	servers, err := mc.ListMirrorServers(limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(servers)
}

func (h *handlers) viewMirrorServer(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	mc, _, err := h.resolveMirrorClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	server, err := mc.GetMirrorServer(id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(server)
}

func (h *handlers) listMirroredRepos(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	mirrorID, err := requireString(req, "mirror_id")
	if err != nil {
		return errResultErr(err), nil
	}
	mc, _, err := h.resolveMirrorClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 30)
	repos, err := mc.ListMirroredRepos(mirrorID, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repos)
}
