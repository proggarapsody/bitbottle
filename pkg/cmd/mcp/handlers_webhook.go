package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWebhooks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
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
	hooks, err := client.ListWebhooks(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hooks)
}

func (h *handlers) getWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hook, err := client.GetWebhook(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hook)
}

func (h *handlers) createWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	url, err := requireString(req, "url")
	if err != nil {
		return errResultErr(err), nil
	}
	events := req.GetStringSlice("events", nil)
	if len(events) == 0 {
		return errResult("events: required and must contain at least one event key"), nil
	}
	active := req.GetBool("active", true)
	secret := req.GetString("secret", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hook, err := client.CreateWebhook(project, slug, backend.CreateWebhookInput{
		URL:    url,
		Events: events,
		Active: active,
		Secret: secret,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hook)
}

func (h *handlers) deleteWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeleteWebhook(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"id": id, "status": "deleted"})
}
