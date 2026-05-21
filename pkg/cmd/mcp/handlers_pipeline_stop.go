package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) stopPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := pc.StopPipeline(ns, slug, uuid); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("Stopped pipeline " + uuid), nil
}
