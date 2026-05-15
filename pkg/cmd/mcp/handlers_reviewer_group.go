package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveReviewerGroupClient is the shared preamble for all reviewer-group
// handlers: parse hostname + repo, dial backend, type-assert ReviewerGroupClient.
func (h *handlers) resolveReviewerGroupClient(req mcplib.CallToolRequest) (backend.ReviewerGroupClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return nil, "", "", err
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	rg, err := backend.AsReviewerGroupClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return rg, ns, slug, nil
}

func (h *handlers) listReviewerGroups(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	rg, ns, slug, err := h.resolveReviewerGroupClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	groups, err := rg.ListReviewerGroups(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(groups)
}

func (h *handlers) addReviewerGroup(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	usersRaw, err := requireString(req, "users")
	if err != nil {
		return errResultErr(err), nil
	}
	userSlugs := splitUsersParam(usersRaw)
	requiredApprovals := req.GetInt("required_approvals", 1)
	if requiredApprovals <= 0 {
		requiredApprovals = 1
	}

	rg, ns, slug, err := h.resolveReviewerGroupClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	group, err := rg.CreateReviewerGroup(ns, slug, backend.CreateReviewerGroupInput{
		Name:              name,
		UserSlugs:         userSlugs,
		RequiredApprovals: requiredApprovals,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(group)
}

func (h *handlers) removeReviewerGroup(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id <= 0 {
		return errResult("missing required parameter: id"), nil
	}
	rg, ns, slug, err := h.resolveReviewerGroupClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rg.DeleteReviewerGroup(ns, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "removed"})
}

// splitUsersParam splits a comma-separated user slug string, trimming whitespace.
func splitUsersParam(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
