package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── get_rate_limit_config ─────────────────────────────────────────────────────

func (h *handlers) getRateLimitConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := ac.GetRateLimitConfig()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cfg)
}

// ── set_rate_limit_config ─────────────────────────────────────────────────────

func (h *handlers) setRateLimitConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}

	args := req.GetArguments()
	_, enabledProvided := args["enabled"]
	_, rphProvided := args["requests_per_hour"]
	_, twProvided := args["throttle_wait_ms"]
	if !enabledProvided && !rphProvided && !twProvided {
		return errResult("at least one of enabled, requests_per_hour, or throttle_wait_ms must be provided"), nil
	}

	// Fetch current config so we can merge partial updates.
	current, err := ac.GetRateLimitConfig()
	if err != nil {
		return errResultErr(err), nil
	}

	in := current
	if enabledProvided {
		in.Enabled = req.GetBool("enabled", current.Enabled)
	}
	if rphProvided {
		in.RequestsPerHour = req.GetInt("requests_per_hour", current.RequestsPerHour)
	}
	if twProvided {
		in.ThrottleWaitMS = req.GetInt("throttle_wait_ms", current.ThrottleWaitMS)
	}

	if err := ac.SetRateLimitConfig(in); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(backend.RateLimitConfig{
		Enabled:         in.Enabled,
		RequestsPerHour: in.RequestsPerHour,
		ThrottleWaitMS:  in.ThrottleWaitMS,
	})
}
