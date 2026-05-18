package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

func (h *handlers) listPipelineVariables(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
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
	ops, err := shared.ResolveVariableOps("repository", client, hostname, project, slug, "")
	if err != nil {
		return errResultErr(err), nil
	}
	vars, err := ops.ListVariables()
	if err != nil {
		return errResultErr(err), nil
	}
	// Defensive: even though the wire layer never returns Value for secured
	// vars, blank it again before serialising in case a future API change
	// leaks one through.
	for i := range vars {
		if vars[i].Secured {
			vars[i].Value = ""
		}
	}
	return jsonResult(vars)
}

func (h *handlers) setPipelineVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	value, err := requireString(req, "value")
	if err != nil {
		return errResultErr(err), nil
	}
	secured := req.GetBool("secured", false)
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ops, err := shared.ResolveVariableOps("repository", client, hostname, project, slug, "")
	if err != nil {
		return errResultErr(err), nil
	}
	v, err := ops.SetVariable(key, value, secured)
	if err != nil {
		return errResultErr(err), nil
	}
	if v.Secured {
		v.Value = ""
	}
	return jsonResult(v)
}

func (h *handlers) deletePipelineVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ops, err := shared.ResolveVariableOps("repository", client, hostname, project, slug, "")
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ops.DeleteVariableByKey(key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"key": key, "status": "deleted"})
}
