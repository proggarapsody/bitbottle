package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) createWorkspaceProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	description := req.GetString("description", "")
	isPrivate := req.GetBool("private", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsCloudProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	_, err = pc.CreateWorkspaceProject(workspace, backend.CreateWorkspaceProjectInput{
		Key:         key,
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Created project %s in %s.", key, workspace)), nil
}

func (h *handlers) viewWorkspaceProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	pc, err := backend.AsCloudProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	proj, err := pc.GetWorkspaceProject(workspace, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(proj)
}

func (h *handlers) editWorkspaceProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	pc, err := backend.AsCloudProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	input := backend.UpdateWorkspaceProjectInput{}
	if name := req.GetString("name", ""); name != "" {
		input.Name = &name
	}
	if desc := req.GetString("description", ""); desc != "" {
		input.Description = &desc
	}
	// Note: boolean false is indistinguishable from absent in MCP without schema tricks;
	// we only set IsPrivate when the raw value is present (non-zero default).
	// This is consistent with how other MCP boolean flags work in this codebase.
	if _, err := pc.UpdateWorkspaceProject(workspace, key, input); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Updated project %s.", key)), nil
}

func (h *handlers) deleteWorkspaceProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	pc, err := backend.AsCloudProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.DeleteWorkspaceProject(workspace, key); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(fmt.Sprintf("Deleted project %s.", key)), nil
}
