package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveBranchRuleClient is the shared preamble for all branch-rule handlers:
// parse hostname + repo, dial backend, type-assert BranchRuleClient.
func (h *handlers) resolveBranchRuleClient(req mcplib.CallToolRequest) (backend.BranchRuleClient, string, string, error) {
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
	br, err := backend.AsBranchRuleClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return br, ns, slug, nil
}

func (h *handlers) listBranchRules(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	br, ns, slug, err := h.resolveBranchRuleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	rules, err := br.ListBranchRules(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(rules)
}

func (h *handlers) addBranchRule(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	kind, err := requireString(req, "kind")
	if err != nil {
		return errResultErr(err), nil
	}
	pattern, err := requireString(req, "pattern")
	if err != nil {
		return errResultErr(err), nil
	}
	value := req.GetInt("value", 0)
	br, ns, slug, err := h.resolveBranchRuleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	added, err := br.AddBranchRule(ns, slug, backend.BranchRuleInput{
		Kind:    kind,
		Pattern: pattern,
		Value:   value,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(added)
}

func (h *handlers) deleteBranchRule(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id <= 0 {
		return errResult("missing required parameter: id"), nil
	}
	br, ns, slug, err := h.resolveBranchRuleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := br.DeleteBranchRule(ns, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "deleted"})
}

func (h *handlers) updateBranchRule(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id <= 0 {
		return errResult("missing required parameter: id"), nil
	}
	br, ns, slug, err := h.resolveBranchRuleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	var in backend.UpdateBranchRuleInput
	if pattern := req.GetString("pattern", ""); pattern != "" {
		in.Pattern = &pattern
	}
	if usersStr := req.GetString("users", ""); usersStr != "" {
		us := strings.Split(usersStr, ",")
		in.Users = &us
	}
	if groupsStr := req.GetString("groups", ""); groupsStr != "" {
		gs := strings.Split(groupsStr, ",")
		in.Groups = &gs
	}
	// Check if value was explicitly provided. Default sentinel is -1 (impossible valid value);
	// value >= 0 means explicitly set — this correctly allows value=0 to be sent.
	if value := req.GetInt("value", -1); value >= 0 {
		in.Value = &value
	}
	updated, err := br.UpdateBranchRule(ns, slug, id, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(updated)
}
