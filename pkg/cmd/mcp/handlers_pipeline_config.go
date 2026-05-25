package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePipelineConfigClient is the shared preamble for all pipeline-config handlers:
// parse hostname, project, slug, dial backend, type-assert PipelineConfigClient.
func (h *handlers) resolvePipelineConfigClient(req mcplib.CallToolRequest) (backend.PipelineConfigClient, string, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	pc, err := backend.AsPipelineConfigClient(client, hostname)
	if err != nil {
		return nil, "", "", "", err
	}
	return pc, project, slug, hostname, nil
}

func (h *handlers) getPipelineConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	pc, project, slug, _, err := h.resolvePipelineConfigClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := pc.GetPipelinesConfig(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cfg)
}

func (h *handlers) enablePipelines(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	pc, project, slug, _, err := h.resolvePipelineConfigClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if _, err := pc.UpdatePipelinesConfig(project, slug, backend.PipelineConfig{Enabled: true}); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Pipelines enabled for %s/%s.", project, slug)), nil
}

func (h *handlers) disablePipelines(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	pc, project, slug, _, err := h.resolvePipelineConfigClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if _, err := pc.UpdatePipelinesConfig(project, slug, backend.PipelineConfig{Enabled: false}); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Pipelines disabled for %s/%s.", project, slug)), nil
}
