package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveBranchModelClient is the shared preamble for all branch-model handlers.
func (h *handlers) resolveBranchModelClient(req mcplib.CallToolRequest) (backend.BranchModelClient, string, string, error) {
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
	bm, err := backend.AsBranchModelClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return bm, ns, slug, nil
}

func (h *handlers) getBranchModel(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	bm, ns, slug, err := h.resolveBranchModelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	model, err := bm.GetBranchModel(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(model)
}

func (h *handlers) setBranchModel(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	bm, ns, slug, err := h.resolveBranchModelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}

	// Fetch current settings to merge branch type prefixes.
	current, err := bm.GetBranchModelSettings(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}

	in := backend.BranchModelSettingsInput{}

	if devBranch := req.GetString("dev_branch", ""); devBranch != "" {
		in.Development = &backend.BranchModelSettingsBranch{
			Name:          devBranch,
			UseMainbranch: false,
		}
	}

	if prodBranch := req.GetString("prod_branch", ""); prodBranch != "" {
		prodEnabled := req.GetBool("prod_enabled", false)
		in.Production = &backend.BranchModelSettingsBranch{
			Name:          prodBranch,
			UseMainbranch: false,
			IsValid:       prodEnabled,
		}
	}

	// branch_type_prefixes is an optional object map[string]string.
	args, _ := req.Params.Arguments.(map[string]any)
	if prefixesRaw, ok := args["branch_type_prefixes"]; ok && prefixesRaw != nil {
		prefixMap, castOK := prefixesRaw.(map[string]any)
		if !castOK {
			return errResult("branch_type_prefixes must be an object"), nil
		}
		overrides := make(map[string]string, len(prefixMap))
		for k, v := range prefixMap {
			s, sOK := v.(string)
			if !sOK {
				return errResult(fmt.Sprintf("branch_type_prefixes[%q] must be a string", k)), nil
			}
			overrides[k] = s
		}
		merged := make([]backend.BranchTypeSettings, 0, len(current.BranchTypes))
		for _, bt := range current.BranchTypes {
			if prefix, ok := overrides[bt.Kind]; ok {
				bt.Prefix = prefix
				delete(overrides, bt.Kind)
			}
			merged = append(merged, bt)
		}
		for kind, prefix := range overrides {
			merged = append(merged, backend.BranchTypeSettings{
				Enabled: true,
				Kind:    kind,
				Prefix:  prefix,
			})
		}
		in.BranchTypes = merged
	}

	updated, err := bm.UpdateBranchModelSettings(ns, slug, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(updated)
}
