package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) listCommits(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	branch := req.GetString("branch", "main")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	commits, err := client.ListCommits(project, slug, branch, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commits)
}

func (h *handlers) getCommit(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	commit, err := client.GetCommit(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commit)
}

func (h *handlers) listCommitStatuses(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	statuses, err := client.ListCommitStatuses(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(statuses)
}
