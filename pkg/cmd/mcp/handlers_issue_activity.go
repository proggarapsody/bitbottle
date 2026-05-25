package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listIssueActivity(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	repoSlug, err := requireString(req, "repo_slug")
	if err != nil {
		return errResultErr(err), nil
	}
	issueID := req.GetInt("issue_id", 0)
	if issueID == 0 {
		return errResult("missing required parameter: issue_id"), nil
	}
	limit := req.GetInt("limit", 25)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	ac, err := backend.AsIssueActivityClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	changes, err := ac.ListIssueActivity(workspace, repoSlug, issueID, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(changes)
}
