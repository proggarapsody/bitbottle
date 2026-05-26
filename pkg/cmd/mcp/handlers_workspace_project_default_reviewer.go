package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveProjectDefaultReviewerClient is the shared preamble for project-default-reviewer handlers.
func (h *handlers) resolveProjectDefaultReviewerClient(req mcplib.CallToolRequest) (backend.WorkspaceProjectDefaultReviewerClient, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, err
	}
	rc, err := backend.AsWorkspaceProjectDefaultReviewerClient(client, hostname)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (h *handlers) listProjectDefaultReviewers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 0)
	rc, err := h.resolveProjectDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	reviewers, err := rc.ListProjectDefaultReviewers(ws, projectKey, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reviewers)
}

func (h *handlers) addProjectDefaultReviewer(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := h.resolveProjectDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rc.AddProjectDefaultReviewer(ws, projectKey, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{
		"workspace":   ws,
		"project_key": projectKey,
		"user":        user,
		"status":      "added",
	})
}

func (h *handlers) removeProjectDefaultReviewer(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	rc, err := h.resolveProjectDefaultReviewerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rc.RemoveProjectDefaultReviewer(ws, projectKey, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{
		"workspace":   ws,
		"project_key": projectKey,
		"user":        user,
		"status":      "removed",
	})
}
