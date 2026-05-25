package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listWorkspacePipelineVars(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	wpc, err := backend.AsWorkspacePipelineVariableClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	vars, err := wpc.ListWorkspacePipelineVariables(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	for i := range vars {
		if vars[i].Secured {
			vars[i].Value = ""
		}
	}
	return jsonResult(vars)
}

func (h *handlers) getWorkspacePipelineVar(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
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

	wpc, err := backend.AsWorkspacePipelineVariableClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	// Resolve key → UUID via list.
	vars, err := wpc.ListWorkspacePipelineVariables(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	var uuid string
	for _, v := range vars {
		if v.Key == key {
			uuid = v.UUID
			break
		}
	}
	if uuid == "" {
		return errResult(fmt.Sprintf("workspace pipeline variable %q not found", key)), nil
	}

	v, err := wpc.GetWorkspacePipelineVariable(workspace, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	if v.Secured {
		v.Value = ""
	}
	return jsonResult(v)
}

func (h *handlers) setWorkspacePipelineVar(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
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

	wpc, err := backend.AsWorkspacePipelineVariableClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	v, err := wpc.SetWorkspacePipelineVariable(workspace, backend.PipelineVariableInput{
		Key:     key,
		Value:   value,
		Secured: secured,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	if v.Secured {
		v.Value = ""
	}
	return jsonResult(v)
}

func (h *handlers) deleteWorkspacePipelineVar(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
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

	wpc, err := backend.AsWorkspacePipelineVariableClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	// Resolve key → UUID via list.
	vars, err := wpc.ListWorkspacePipelineVariables(workspace)
	if err != nil {
		return errResultErr(err), nil
	}
	var uuid string
	for _, v := range vars {
		if v.Key == key {
			uuid = v.UUID
			break
		}
	}
	if uuid == "" {
		return errResult(fmt.Sprintf("workspace pipeline variable %q not found", key)), nil
	}

	if err := wpc.DeleteWorkspacePipelineVariable(workspace, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"key": key, "status": "deleted"})
}
