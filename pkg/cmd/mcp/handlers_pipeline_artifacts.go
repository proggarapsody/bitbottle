package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

const artifactSizeWarningBytes = 5 * 1024 * 1024 // 5 MB

func (h *handlers) listPipelineArtifacts(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 50)
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
	ac, err := backend.AsPipelineArtifactClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	artifacts, err := ac.ListPipelineArtifacts(project, slug, pipelineUUID, stepUUID, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(artifacts)
}

func (h *handlers) downloadPipelineArtifact(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}

	// Check artifact size first by listing artifacts.
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ac, err := backend.AsPipelineArtifactClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	// List artifacts to find the size of the requested artifact.
	artifacts, err := ac.ListPipelineArtifacts(project, slug, pipelineUUID, stepUUID, 100)
	if err != nil {
		return errResultErr(err), nil
	}
	for _, a := range artifacts {
		if a.Name == name && a.SizeBytes > artifactSizeWarningBytes {
			return errResult(fmt.Sprintf(
				"artifact %q is %.1f MB; use `pipeline artifact download --out PATH` instead",
				name, float64(a.SizeBytes)/float64(1024*1024),
			)), nil
		}
	}

	var buf bytes.Buffer
	if err := ac.DownloadPipelineArtifact(project, slug, pipelineUUID, stepUUID, name, &buf); err != nil {
		return errResultErr(err), nil
	}

	return jsonResult(map[string]any{
		"name":           name,
		"content_base64": base64.StdEncoding.EncodeToString(buf.Bytes()),
	})
}
