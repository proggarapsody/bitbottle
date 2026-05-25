package mcp

import (
	"context"
	"fmt"
	"strings"

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

// ── list_admin_users ──────────────────────────────────────────────────────────

func (h *handlers) listAdminUsers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	filter := req.GetString("filter", "")
	limit := req.GetInt("limit", 50)
	if err := validateRange("limit", limit, 1, 1000); err != nil {
		return errResult(err.Error()), nil
	}
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	users, err := ac.ListAdminUsers(filter, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(users)
}

// ── activate_user ─────────────────────────────────────────────────────────────

func (h *handlers) activateUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.ActivateUser(slug); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"slug": slug, "status": "activated"})
}

// ── deactivate_user ───────────────────────────────────────────────────────────

func (h *handlers) deactivateUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.DeactivateUser(slug); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"slug": slug, "status": "deactivated"})
}

// ── rename_user ───────────────────────────────────────────────────────────────

func (h *handlers) renameUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	newSlug, err := requireString(req, "new_slug")
	if err != nil {
		return errResultErr(err), nil
	}
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.RenameUser(slug, newSlug); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"old_slug": slug, "new_slug": newSlug, "status": "renamed"})
}

// ── get_admin_license ─────────────────────────────────────────────────────────

func (h *handlers) getAdminLicense(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	lic, err := ac.GetLicense()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(lic)
}

// ── get_cluster_nodes ─────────────────────────────────────────────────────────

func (h *handlers) getClusterNodes(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	nodes, err := ac.GetClusterNodes()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(nodes)
}

// ── get_mail_server_config ────────────────────────────────────────────────────

func (h *handlers) getMailServerConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := ac.GetMailServerConfig()
	if err != nil {
		return errResultErr(err), nil
	}
	// Return the domain struct directly — Password has json:"-" so it is
	// excluded from the JSON output automatically.
	return jsonResult(cfg)
}

// ── set_mail_server_config ────────────────────────────────────────────────────

func (h *handlers) setMailServerConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	mailHostname, err := requireString(req, "mail_hostname")
	if err != nil {
		return errResultErr(err), nil
	}
	port := req.GetInt("port", 25)
	protocol := req.GetString("protocol", "smtp")
	useStartTLS := req.GetBool("use_starttls", false)
	requireStartTLS := req.GetBool("require_starttls", false)
	username := req.GetString("username", "")
	senderAddress := req.GetString("sender_address", "")
	password := req.GetString("password", "")

	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.SetMailServerConfig(backend.MailServerConfig{
		Hostname:        mailHostname,
		Port:            port,
		Protocol:        protocol,
		UseStartTLS:     useStartTLS,
		RequireStartTLS: requireStartTLS,
		Username:        username,
		SenderAddress:   senderAddress,
		Password:        password,
	}); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "updated", "mail_hostname": mailHostname})
}

// ── get_banner ────────────────────────────────────────────────────────────────

func (h *handlers) getBanner(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := ac.GetBanner()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cfg)
}

// ── set_banner ────────────────────────────────────────────────────────────────

func (h *handlers) setBanner(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	message, err := requireString(req, "message")
	if err != nil {
		return errResultErr(err), nil
	}
	audience := strings.ToUpper(req.GetString("audience", "ALL"))
	switch audience {
	case "ALL", "AUTHENTICATED", "UNAUTHENTICATED":
	default:
		return errResult(fmt.Sprintf("audience must be one of ALL, AUTHENTICATED, UNAUTHENTICATED; got %q", audience)), nil
	}
	enabled := req.GetBool("enabled", true)

	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.SetBanner(backend.BannerConfig{
		Message:  message,
		Audience: audience,
		Enabled:  enabled,
	}); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "updated", "message": message})
}

// ── clear_banner ──────────────────────────────────────────────────────────────

func (h *handlers) clearBanner(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ac, _, err := h.resolveAdminClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ac.ClearBanner(); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "cleared", "message": "Site-wide announcement banner has been removed."})
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
