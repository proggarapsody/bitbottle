package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveCommitFileClient is the shared preamble for commit-file handlers:
// parse hostname + repo, dial backend, type-assert CommitFileClient.
func (h *handlers) resolveCommitFileClient(req mcplib.CallToolRequest) (backend.CommitFileClient, string, string, error) {
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
	cf, err := backend.AsCommitFileClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return cf, ns, slug, nil
}

func (h *handlers) listCommitFiles(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	cf, ns, slug, err := h.resolveCommitFileClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	entries, err := cf.ListCommitFiles(ns, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(entries)
}
