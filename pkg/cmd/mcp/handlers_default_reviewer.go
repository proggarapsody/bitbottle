package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveDefaultReviewerClient is the shared preamble for all default-reviewer handlers:
// parse hostname + repo, dial backend, type-assert DefaultReviewerClient.
func (h *handlers) resolveDefaultReviewerClient(req mcplib.CallToolRequest) (backend.DefaultReviewerClient, string, string, error) {
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
	dr, err := backend.AsDefaultReviewerClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return dr, ns, slug, nil
}

func (h *handlers) listDefaultReviewers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	dr, ns, slug, err := h.resolveDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	reviewers, err := dr.ListDefaultReviewers(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reviewers)
}

func (h *handlers) addDefaultReviewer(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	dr, ns, slug, err := h.resolveDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := dr.AddDefaultReviewer(ns, slug, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"user": user, "status": "added"})
}

func (h *handlers) removeDefaultReviewer(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	dr, ns, slug, err := h.resolveDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := dr.RemoveDefaultReviewer(ns, slug, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"user": user, "status": "removed"})
}
