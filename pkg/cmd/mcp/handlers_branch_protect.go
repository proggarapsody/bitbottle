package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveBranchProtector mirrors resolveIssueClient: pick the host, dial the
// backend, type-assert BranchProtector, and gather the project/slug args.
func (h *handlers) resolveBranchProtector(req mcplib.CallToolRequest) (backend.BranchProtector, string, string, error) {
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
	bp, err := backend.AsBranchProtector(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return bp, project, slug, nil
}

func (h *handlers) listBranchProtections(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}
	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	got, err := bp.ListBranchProtections(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(got)
}

func (h *handlers) createBranchProtection(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	typ, err := requireString(req, "type")
	if err != nil {
		return errResultErr(err), nil
	}
	branch := req.GetString("branch", "")
	pattern := req.GetString("pattern", "")
	if (branch == "") == (pattern == "") {
		return errResult("specify exactly one of branch or pattern"), nil
	}
	matcherID := branch
	matcherKind := "BRANCH"
	if pattern != "" {
		matcherID = pattern
		matcherKind = "PATTERN"
	}
	users := req.GetStringSlice("users", nil)
	groups := req.GetStringSlice("groups", nil)

	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	out, err := bp.CreateBranchProtection(project, slug, backend.CreateBranchProtectionInput{
		Type:        typ,
		MatcherID:   matcherID,
		MatcherKind: matcherKind,
		Users:       users,
		Groups:      groups,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(out)
}

func (h *handlers) deleteBranchProtection(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := bp.DeleteBranchProtection(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "deleted"})
}
