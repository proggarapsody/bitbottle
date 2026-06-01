package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveRepoHookClient resolves the backend and type-asserts to
// RepoHookClient, returning a host.unsupported error on Cloud.
func (h *handlers) resolveRepoHookClient(req mcplib.CallToolRequest) (backend.RepoHookClient, string, string, error) {
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
	hc, err := backend.AsRepoHookClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return hc, project, slug, nil
}

func (h *handlers) listRepoHooks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hooks, err := hc.ListRepoHooks(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hooks)
}

func (h *handlers) viewRepoHook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hookKey, err := requireString(req, "hook_key")
	if err != nil {
		return errResultErr(err), nil
	}
	hook, err := hc.GetRepoHook(project, slug, hookKey)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hook)
}

func (h *handlers) enableRepoHook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hookKey, err := requireString(req, "hook_key")
	if err != nil {
		return errResultErr(err), nil
	}
	if err := hc.EnableRepoHook(project, slug, hookKey); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "enabled", "key": hookKey})
}

func (h *handlers) disableRepoHook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hookKey, err := requireString(req, "hook_key")
	if err != nil {
		return errResultErr(err), nil
	}
	if err := hc.DisableRepoHook(project, slug, hookKey); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "disabled", "key": hookKey})
}

func (h *handlers) getRepoHookSettings(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hookKey, err := requireString(req, "hook_key")
	if err != nil {
		return errResultErr(err), nil
	}
	raw, err := hc.GetRepoHookSettings(project, slug, hookKey)
	if err != nil {
		return errResultErr(err), nil
	}
	// Return the raw JSON as a text result so callers can pipe it directly.
	return mcplib.NewToolResultText(string(raw)), nil
}

func (h *handlers) setRepoHookSettings(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hc, project, slug, err := h.resolveRepoHookClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	hookKey, err := requireString(req, "hook_key")
	if err != nil {
		return errResultErr(err), nil
	}
	configStr, err := requireString(req, "config")
	if err != nil {
		return errResultErr(err), nil
	}
	var cfg json.RawMessage
	if jerr := json.Unmarshal([]byte(configStr), &cfg); jerr != nil {
		return errResult(fmt.Sprintf("invalid config JSON: %v", jerr)), nil
	}
	if err := hc.SetRepoHookSettings(project, slug, hookKey, cfg); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "ok", "key": hookKey})
}
