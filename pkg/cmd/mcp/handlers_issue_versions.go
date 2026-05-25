package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listIssueVersions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	vc, err := backend.AsIssueVersionClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	versions, err := vc.ListIssueVersions(workspace, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(versions)
}

func (h *handlers) viewIssueVersion(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("required parameter missing or zero: id"), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	vc, err := backend.AsIssueVersionClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	v, err := vc.GetIssueVersion(workspace, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(v)
}

func (h *handlers) createIssueVersion(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	vc, err := backend.AsIssueVersionClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	v, err := vc.CreateIssueVersion(workspace, slug, name)
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Created version %q (ID: %d).", v.Name, v.ID)), nil
}

func (h *handlers) deleteIssueVersion(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("required parameter missing or zero: id"), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	vc, err := backend.AsIssueVersionClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := vc.DeleteIssueVersion(workspace, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Deleted version %d.", id)), nil
}
