package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listMilestones(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	project, err := requireString(req, "project")
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
	mc, err := backend.AsMilestoneClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	milestones, err := mc.ListMilestones(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(milestones)
}

func (h *handlers) viewMilestone(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	idFloat := req.GetInt("id", 0)
	if idFloat == 0 {
		return errResult(fmt.Sprintf("required parameter missing or zero: id")), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	mc, err := backend.AsMilestoneClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	m, err := mc.GetMilestone(project, slug, idFloat)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(m)
}
