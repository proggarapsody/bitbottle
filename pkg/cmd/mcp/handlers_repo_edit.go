package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) editRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}

	// Build input from optional fields. We check the raw arguments map so that
	// an explicitly-supplied false boolean is distinguished from an absent field.
	args, _ := req.Params.Arguments.(map[string]any)

	in := backend.EditRepoInput{}
	noop := true

	if desc := req.GetString("description", ""); desc != "" {
		in.Description = &desc
		noop = false
	} else if _, ok := args["description"]; ok {
		// empty string explicitly set
		empty := ""
		in.Description = &empty
		noop = false
	}

	if website := req.GetString("website", ""); website != "" {
		in.Website = &website
		noop = false
	}
	if language := req.GetString("language", ""); language != "" {
		in.Language = &language
		noop = false
	}
	if forkPolicy := req.GetString("fork_policy", ""); forkPolicy != "" {
		in.ForkPolicy = &forkPolicy
		noop = false
	}

	if _, ok := args["has_issues"]; ok {
		v := req.GetBool("has_issues", false)
		in.HasIssues = &v
		noop = false
	}
	if _, ok := args["has_wiki"]; ok {
		v := req.GetBool("has_wiki", false)
		in.HasWiki = &v
		noop = false
	}

	if noop {
		return errResult("no fields to update; supply at least one of: description, website, language, fork_policy, has_issues, has_wiki"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	editor, err := backend.AsRepoEditor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	updated, err := editor.EditRepo(ns, slug, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(updated)
}
