package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspaceHooks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wwc, err := backend.AsWorkspaceWebhookClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	hooks, err := wwc.ListWorkspaceWebhooks(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hooks)
}

func (h *handlers) createWorkspaceHook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	hookURL, err := requireString(req, "url")
	if err != nil {
		return errResultErr(err), nil
	}
	eventsRaw, err := requireString(req, "events")
	if err != nil {
		return errResultErr(err), nil
	}
	active := req.GetBool("active", true)

	var events []string
	for _, e := range strings.Split(eventsRaw, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			events = append(events, e)
		}
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wwc, err := backend.AsWorkspaceWebhookClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	hook, err := wwc.CreateWorkspaceWebhook(workspace, backend.CreateWebhookInput{
		URL:    hookURL,
		Events: events,
		Active: active,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Created workspace webhook %s.", hook.ID)), nil
}

func (h *handlers) deleteWorkspaceHook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wwc, err := backend.AsWorkspaceWebhookClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := wwc.DeleteWorkspaceWebhook(workspace, uuid); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Deleted workspace webhook %s.", uuid)), nil
}
