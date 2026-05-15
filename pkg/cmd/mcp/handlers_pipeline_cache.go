package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePipelineCacheClient is the shared preamble for all pipeline-cache
// handlers: parse hostname + repo, dial backend, type-assert PipelineCacheClient.
func (h *handlers) resolvePipelineCacheClient(req mcplib.CallToolRequest) (backend.PipelineCacheClient, string, string, error) {
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
	pc, err := backend.AsPipelineCacheClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return pc, ns, slug, nil
}

func (h *handlers) listPipelineCaches(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	pc, ns, slug, err := h.resolvePipelineCacheClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	caches, err := pc.ListPipelineCaches(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(caches)
}

func (h *handlers) deletePipelineCache(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	pc, ns, slug, err := h.resolvePipelineCacheClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.DeletePipelineCache(ns, slug, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"uuid": uuid, "status": "deleted"})
}
