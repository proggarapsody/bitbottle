package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) syncRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	ns, err := requireString(req, "ns")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	branch := req.GetString("branch", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	syncer, err := backend.AsRepoSyncer(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	result, err := syncer.SyncRepo(ns, slug, branch)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(result)
}
