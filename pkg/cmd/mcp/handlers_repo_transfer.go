package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) transferRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	target, err := requireString(req, "target")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	rt, err := backend.AsRepoTransferClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	updated, err := rt.TransferRepo(ns, slug, target)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(updated)
}
