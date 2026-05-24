package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// scopeAliasesMCP mirrors the CLI's scope alias map for the MCP layer.
var scopeAliasesMCP = map[string]string{
	"repo:read":     "REPO_READ",
	"repo:write":    "REPO_WRITE",
	"pr:read":       "PR_READ",
	"pr:write":      "PR_WRITE",
	"project:read":  "PROJECT_READ",
	"project:write": "PROJECT_WRITE",
}

// resolvePATBackend resolves the backend client and userSlug for PAT operations.
func (h *handlers) resolvePATBackend(hostname string) (backend.PATClient, string, error) {
	host, err := factory.ResolveHost(h.f, hostname)
	if err != nil {
		return nil, "", err
	}

	cfg, err := h.f.Config()
	if err != nil {
		return nil, "", err
	}
	hc, ok := cfg.Get(host)
	if !ok {
		return nil, "", fmt.Errorf("not logged into %s", host)
	}
	userSlug := hc.User
	if userSlug == "" {
		return nil, "", fmt.Errorf("no username stored for host %s; re-run `bitbottle auth login`", host)
	}

	client, err := h.f.Backend(host)
	if err != nil {
		return nil, "", err
	}

	pc, err := backend.AsPATClient(client, host)
	if err != nil {
		return nil, "", err
	}
	return pc, userSlug, nil
}

func (h *handlers) listPATs(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 50)

	pc, userSlug, err := h.resolvePATBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	pats, err := pc.ListPATs(userSlug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pats)
}

// createPAT creates a PAT via the MCP tool.
// IMPORTANT: the response includes a "token" field containing the raw secret.
// The secret is returned exactly once and cannot be retrieved again.
func (h *handlers) createPAT(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	name, err := requireString(req, "name")
	if err != nil {
		return errResult(err.Error()), nil
	}
	scopesRaw, err := requireString(req, "scopes")
	if err != nil {
		return errResult(err.Error()), nil
	}

	permissions, err := resolveMCPScopes(scopesRaw)
	if err != nil {
		return errResult(err.Error()), nil
	}

	pc, userSlug, err := h.resolvePATBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	in := backend.CreatePATInput{
		Name:        name,
		Permissions: permissions,
	}
	if expiresIn := req.GetInt("expires_in", 0); expiresIn > 0 {
		in.ExpiryDays = &expiresIn
	}

	pat, err := pc.CreatePAT(userSlug, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pat)
}

func (h *handlers) revokePAT(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	tokenID, err := requireString(req, "token_id")
	if err != nil {
		return errResult(err.Error()), nil
	}

	pc, userSlug, err := h.resolvePATBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := pc.RevokePAT(userSlug, tokenID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "revoked", "token_id": tokenID})
}

// resolveMCPScopes converts a comma-separated scope string to Bitbucket permission names.
func resolveMCPScopes(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if canon, ok := scopeAliasesMCP[p]; ok {
			out = append(out, canon)
			continue
		}
		upper := strings.ToUpper(p)
		valid := false
		for _, v := range scopeAliasesMCP {
			if v == upper {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("unknown scope %q; valid: repo:read, repo:write, pr:read, pr:write, project:read, project:write", p)
		}
		out = append(out, upper)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scopes must not be empty")
	}
	return out, nil
}
