package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspaceIPAllowlists(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ac, err := backend.AsIPAllowlistClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	entries, err := ac.ListIPAllowlists(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(entries)
}

func (h *handlers) createWorkspaceIPAllowlist(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	cidr, err := requireString(req, "cidr")
	if err != nil {
		return errResultErr(err), nil
	}
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ac, err := backend.AsIPAllowlistClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	enabledStr := req.GetString("enabled", "true")
	enabled := enabledStr != "false"
	entry, err := ac.CreateIPAllowlist(workspace, backend.CreateIPAllowlistInput{
		CIDR:        cidr,
		Description: req.GetString("description", ""),
		Enabled:     enabled,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(entry)
}

func (h *handlers) deleteWorkspaceIPAllowlist(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ac, err := backend.AsIPAllowlistClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.DeleteIPAllowlist(workspace, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "uuid": uuid})
}
