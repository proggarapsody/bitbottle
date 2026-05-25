package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveWorkspacePermsClient is the shared preamble for workspace-perms handlers.
func (h *handlers) resolveWorkspacePermsClient(req mcplib.CallToolRequest) (backend.WorkspacePermsClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	wpc, err := backend.AsWorkspacePermsClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return wpc, hostname, nil
}

func (h *handlers) listWorkspacePerms(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	wpc, _, err := h.resolveWorkspacePermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 50)
	perms, err := wpc.ListWorkspaceMemberPerms(ws, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(perms)
}

func (h *handlers) listWorkspaceRepoPerms(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	wpc, _, err := h.resolveWorkspacePermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 50)
	perms, err := wpc.ListWorkspaceRepoPerms(ws, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(perms)
}

func (h *handlers) grantWorkspacePerm(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	permission, err := requireString(req, "permission")
	if err != nil {
		return errResultErr(err), nil
	}
	wpc, _, err := h.resolveWorkspacePermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := wpc.GrantWorkspacePerm(ws, user, permission); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"workspace": ws, "user": user, "permission": permission, "status": "granted"})
}

func (h *handlers) revokeWorkspacePerm(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}
	wpc, _, err := h.resolveWorkspacePermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := wpc.RevokeWorkspacePerm(ws, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"workspace": ws, "user": user, "status": "revoked"})
}
