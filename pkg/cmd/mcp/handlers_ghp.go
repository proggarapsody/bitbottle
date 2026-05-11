package mcp

import (
	"context"
	"fmt"
	"time"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) prChecks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("host", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
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

	p, err := client.GetPR(project, repo, prID)
	if err != nil {
		return errResultErr(err), nil
	}
	if p.HeadCommitHash == "" {
		return errResult(fmt.Sprintf("head commit hash unavailable for PR #%d", prID)), nil
	}

	statuses, err := client.ListCommitStatuses(project, repo, p.HeadCommitHash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(statuses)
}

func (h *handlers) prUpdateBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("host", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
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

	updater, ok := client.(backend.PRBranchUpdater)
	if !ok {
		return errResult("update-branch is not supported by this backend"), nil
	}

	if err := updater.UpdatePRBranch(project, repo, prID); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(`{"status":"branch updated"}`), nil
}

func (h *handlers) prStatus(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("host", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	lister, ok := client.(backend.PRStatusLister)
	if !ok {
		return errResult("pr status is not supported by this backend"), nil
	}

	entries, err := lister.ListMyPRs(project, repo)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(entries)
}

func (h *handlers) ghpStatus(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	return h.prStatus(context.Background(), req)
}

func (h *handlers) pipelineWatch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("host", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	interval := req.GetInt("interval", 5)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	terminalStates := map[string]bool{
		"COMPLETED": true,
		"FAILED":    true,
		"STOPPED":   true,
		"EXPIRED":   true,
		"ERROR":     true,
	}

	for {
		pl, err := pc.GetPipeline(project, repo, uuid)
		if err != nil {
			return errResultErr(err), nil
		}
		if terminalStates[pl.State] {
			return jsonResult(pl)
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
