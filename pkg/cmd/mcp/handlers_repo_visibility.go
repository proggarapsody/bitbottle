package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) repoVisibility(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
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

	vis := req.GetString("visibility", "")

	// GET mode — no visibility arg
	if vis == "" {
		r, err := client.GetRepo(ns, slug)
		if err != nil {
			return errResultErr(err), nil
		}
		if r.IsPrivate {
			return mcplib.NewToolResultText("private"), nil
		}
		return mcplib.NewToolResultText("public"), nil
	}

	// SET mode
	var isPrivate bool
	switch vis {
	case "public":
		isPrivate = false
	case "private":
		isPrivate = true
	default:
		return errResult(fmt.Sprintf("invalid visibility %q: must be \"public\" or \"private\"", vis)), nil
	}

	if err := client.SetRepoVisibility(ns, slug, isPrivate); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Repository %s/%s is now %s.", ns, slug, vis)), nil
}
