package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) cherryPickCommit(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	commitHash, err := requireString(req, "commit_hash")
	if err != nil {
		return errResultErr(err), nil
	}
	targetBranch, err := requireString(req, "target_branch")
	if err != nil {
		return errResultErr(err), nil
	}
	message := req.GetString("message", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cp, err := backend.AsCommitCherryPicker(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	result, err := cp.CherryPickCommit(ns, slug, backend.CherryPickInput{
		SourceHash:   commitHash,
		TargetBranch: targetBranch,
		Message:      message,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(result)
}
