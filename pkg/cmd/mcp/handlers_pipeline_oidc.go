package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePipelineOIDCClient is the shared preamble for pipeline-oidc handlers:
// parse hostname and workspace, dial backend, type-assert PipelineOIDCClient.
func (h *handlers) resolvePipelineOIDCClient(req mcplib.CallToolRequest) (backend.PipelineOIDCClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	oc, err := backend.AsPipelineOIDCClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return oc, workspace, hostname, nil
}

func (h *handlers) getPipelineOIDCConfig(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	oc, workspace, _, err := h.resolvePipelineOIDCClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	cfg, err := oc.GetPipelineOIDCConfig(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cfg)
}

func (h *handlers) getPipelineOIDCKeys(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	oc, workspace, _, err := h.resolvePipelineOIDCClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	keys, err := oc.GetPipelineOIDCKeys(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(keys)
}
