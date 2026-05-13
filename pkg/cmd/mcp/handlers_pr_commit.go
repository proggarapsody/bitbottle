package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listPRCommits(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	pc, err := backend.AsPRCommitClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	ns = strings.ToUpper(ns)
	commits, err := pc.ListPRCommits(ns, slug, prID)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commits)
}
