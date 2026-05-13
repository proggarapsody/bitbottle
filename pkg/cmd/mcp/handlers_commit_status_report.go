package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) reportCommitStatus(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	state, err := requireString(req, "state")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	input := backend.CommitStatusInput{
		Key:         key,
		State:       state,
		URL:         req.GetString("url", ""),
		Name:        req.GetString("name", ""),
		Description: req.GetString("description", ""),
	}

	status, err := client.ReportCommitStatus(ns, slug, hash, input)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(status)
}
