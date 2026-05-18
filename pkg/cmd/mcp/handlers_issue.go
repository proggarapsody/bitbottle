package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveIssueClient is the small adapter every issue handler shares: pick
// the host, dial the backend, type-assert IssueClient, and gather the
// project/slug args. Keeps each handler body trivial.
func (h *handlers) resolveIssueClient(req mcplib.CallToolRequest) (backend.IssueClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	ic, err := backend.AsIssueClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return ic, project, slug, nil
}

func (h *handlers) listIssues(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	state := req.GetString("state", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issues, err := ic.ListIssues(project, slug, state, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issues)
}

func (h *handlers) getIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.GetIssue(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) createIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	title, err := requireString(req, "title")
	if err != nil {
		return errResultErr(err), nil
	}
	body := req.GetString("body", "")
	kind := req.GetString("kind", "")
	priority := req.GetString("priority", "")
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.CreateIssue(project, slug, backend.CreateIssueInput{
		Title:    title,
		Content:  body,
		Kind:     kind,
		Priority: priority,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) closeIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.UpdateIssue(project, slug, id, backend.UpdateIssueInput{State: "closed"})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) updateIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.UpdateIssueInput{
		Title:    req.GetString("title", ""),
		Content:  req.GetString("body", ""),
		Kind:     req.GetString("kind", ""),
		Priority: req.GetString("priority", ""),
		Assignee: req.GetString("assignee", ""),
		State:    req.GetString("state", ""),
	}
	issue, err := ic.UpdateIssue(project, slug, id, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) reopenIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.ReopenIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "state": "open"})
}

func (h *handlers) assignIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	assignee, err := requireString(req, "assignee")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.AssignIssue(project, slug, id, assignee); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "assignee": assignee})
}
