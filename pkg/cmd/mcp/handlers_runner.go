package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveRunnerClient is the shared preamble for all runner handlers:
// parse hostname, dial backend, type-assert RunnerClient.
func (h *handlers) resolveRunnerClient(req mcplib.CallToolRequest) (backend.RunnerClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	rc, err := backend.AsRunnerClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return rc, hostname, nil
}

// ── list_runners ──────────────────────────────────────────────────────────────

func (h *handlers) listRunners(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	rc, _, err := h.resolveRunnerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	runners, err := rc.ListRunners(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(runners)
}

// ── create_runner ─────────────────────────────────────────────────────────────

var mcpValidPlatforms = map[string]backend.RunnerPlatform{
	"linux_amd64":   {Operating: "LINUX", Arch: "AMD64"},
	"linux_arm64":   {Operating: "LINUX", Arch: "ARM64"},
	"windows_amd64": {Operating: "WINDOWS", Arch: "AMD64"},
	"macos_arm64":   {Operating: "MACOS", Arch: "ARM64"},
}

func (h *handlers) createRunner(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	platformStr := strings.ToLower(req.GetString("platform", "linux_amd64"))
	plat, ok := mcpValidPlatforms[platformStr]
	if !ok {
		return errResult(fmt.Sprintf("invalid platform %q: must be one of linux_amd64, linux_arm64, windows_amd64, macos_arm64", platformStr)), nil
	}

	labels := req.GetStringSlice("labels", nil)

	rc, _, err := h.resolveRunnerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	runner, err := rc.CreateRunner(workspace, backend.CreateRunnerInput{
		Name:     name,
		Labels:   labels,
		Platform: plat,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(runner)
}

// ── delete_runner ─────────────────────────────────────────────────────────────

func (h *handlers) deleteRunner(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	rc, _, err := h.resolveRunnerClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rc.DeleteRunner(workspace, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"uuid": uuid, "status": "deleted"})
}
