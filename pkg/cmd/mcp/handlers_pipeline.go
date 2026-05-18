package mcp

import (
	"context"
	"io"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listPipelines(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 20)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	pipelines, err := pc.ListPipelines(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pipelines)
}

func (h *handlers) getPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	pl, err := pc.GetPipeline(project, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pl)
}

func (h *handlers) runPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	branch, err := requireString(req, "branch")
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
	pl, err := pc.RunPipeline(project, slug, backend.RunPipelineInput{Branch: branch})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pl)
}

func (h *handlers) listPipelineSteps(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	steps, err := pc.ListPipelineSteps(project, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(steps)
}

func (h *handlers) getPipelineStepLog(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	pipelineUUID, err := requireString(req, "pipeline_uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	stepUUID, err := requireString(req, "step_uuid")
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
	rc, err := pc.GetPipelineStepLog(project, slug, pipelineUUID, stepUUID)
	if err != nil {
		return errResultErr(err), nil
	}
	defer rc.Close() //nolint:errcheck
	body, err := io.ReadAll(rc)
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(string(body)), nil
}
