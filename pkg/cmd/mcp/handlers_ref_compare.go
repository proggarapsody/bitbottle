package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) compareRefs(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	base, err := requireString(req, "base")
	if err != nil {
		return errResultErr(err), nil
	}
	head, err := requireString(req, "head")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 30)
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := backend.AsRefComparer(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cmp, err := rc.CompareRefs(ns, slug, base, head, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cmp)
}
