package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) handleSetRepoDefaultBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	branch, err := requireString(req, "branch")
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

	if err := client.SetRepoDefaultBranch(ns, slug, branch); err != nil {
		return errResultErr(err), nil
	}

	return mcplib.NewToolResultText(fmt.Sprintf("Default branch of %s/%s set to %q.", ns, slug, branch)), nil
}
