package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveAdminClient resolves the backend and type-asserts to AdminClient,
// returning a host.unsupported error on Cloud.
func (h *handlers) resolveAdminClient(req mcplib.CallToolRequest) (backend.AdminClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	ac, err := backend.AsAdminClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return ac, hostname, nil
}

// ── rotate_secrets ────────────────────────────────────────────────────────────

func (h *handlers) rotateSecrets(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.RotateSecrets(); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{
		"status":  "rotated",
		"message": "Secrets rotated. Restart all cluster nodes for the new secret to take effect.",
	})
}

// ── get_logging_config ────────────────────────────────────────────────────────

func (h *handlers) getLoggingConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := ac.GetLoggingConfig()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{
		"level": cfg.Level,
		"async": cfg.Async,
	})
}

// ── set_logging_config ────────────────────────────────────────────────────────

func (h *handlers) setLoggingConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	level := req.GetString("level", "")
	asyncVal := req.GetBool("async", false)
	persistentVal := req.GetBool("persistent", false)

	args := req.GetArguments()
	_, asyncProvided := args["async"]
	if level == "" && !asyncProvided {
		return errResult("at least one of level or async must be provided"), nil
	}

	validLevels := map[string]bool{
		"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true,
	}
	if level != "" && !validLevels[level] {
		return errResult(fmt.Sprintf("log level must be one of DEBUG, INFO, WARN, ERROR (case-sensitive), got %q", level)), nil
	}

	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.LoggingConfigInput{
		Level:      level,
		Async:      asyncVal,
		Persistent: persistentVal,
	}
	if err := ac.SetLoggingConfig(in); err != nil {
		return errResultErr(err), nil
	}
	note := "runtime-only change; will reset on next restart"
	if persistentVal {
		note = "persistent change; will survive restarts"
	}
	return jsonResult(map[string]string{"status": "updated", "note": note})
}
