package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) searchCommits(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	opts := backend.CommitSearchOpts{
		Query:  req.GetString("query", ""),
		Author: req.GetString("author", ""),
		Since:  req.GetString("since", ""),
		Until:  req.GetString("until", ""),
		Limit:  req.GetInt("limit", 30),
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cs, err := backend.AsCommitSearcher(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	commits, err := cs.SearchCommits(project, slug, opts)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commits)
}
