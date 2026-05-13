package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveDiffClient is the shared preamble for all diff handlers:
// parse hostname + repo, dial backend, type-assert DiffClient.
func (h *handlers) resolveDiffClient(req mcplib.CallToolRequest) (backend.DiffClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return nil, "", "", err
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	dc, err := backend.AsDiffClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return dc, ns, slug, nil
}

func (h *handlers) getDiff(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	from, err := requireString(req, "from")
	if err != nil {
		return errResultErr(err), nil
	}
	to, err := requireString(req, "to")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDiffClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	text, err := dc.GetDiff(ns, slug, from, to)
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(text), nil
}

func (h *handlers) getDiffStat(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	from, err := requireString(req, "from")
	if err != nil {
		return errResultErr(err), nil
	}
	to, err := requireString(req, "to")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDiffClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	stat, err := dc.GetDiffStat(ns, slug, from, to)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(stat)
}
