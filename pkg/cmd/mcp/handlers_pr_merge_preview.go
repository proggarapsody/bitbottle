package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// dryRunMergePR is the MCP handler for the `dry_run_merge_pr` tool.
// Supported on both Cloud and Server / Data Center via the optional
// PRMergePreviewClient interface; stamps host.unsupported on backends that
// do not implement the interface.
func (h *handlers) dryRunMergePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	strategy := req.GetString("strategy", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	preview, err := backend.AsPRMergePreviewClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	result, err := preview.DryRunMergePR(project, slug, prID, strategy)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(result)
}
