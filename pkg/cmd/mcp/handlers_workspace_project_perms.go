package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveWorkspaceProjectPermsClient is the shared preamble for workspace-project-perms handlers.
func (h *handlers) resolveWorkspaceProjectPermsClient(req mcplib.CallToolRequest) (backend.WorkspaceProjectPermsClient, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, err
	}
	wpc, err := backend.AsWorkspaceProjectPermsClient(client, hostname)
	if err != nil {
		return nil, err
	}
	return wpc, nil
}

func (h *handlers) listWorkspaceProjectPerms(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	wpc, err := h.resolveWorkspaceProjectPermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	perms, err := wpc.ListWorkspaceProjectPerms(ws, projectKey)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(perms)
}

func (h *handlers) grantWorkspaceProjectPerm(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	permission, err := requireString(req, "permission")
	if err != nil {
		return errResultErr(err), nil
	}
	userSlug := req.GetString("user_slug", "")
	groupSlug := req.GetString("group_slug", "")
	if userSlug == "" && groupSlug == "" {
		return errResultErr(fmt.Errorf("user_slug or group_slug is required")), nil
	}
	wpc, err := h.resolveWorkspaceProjectPermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.WorkspaceProjectPermInput{
		Permission: permission,
		UserSlug:   userSlug,
		GroupSlug:  groupSlug,
	}
	if err := wpc.GrantWorkspaceProjectPerm(ws, projectKey, in); err != nil {
		return errResultErr(err), nil
	}
	subject := userSlug
	if groupSlug != "" {
		subject = groupSlug
	}
	return jsonResult(map[string]any{
		"workspace":   ws,
		"project_key": projectKey,
		"subject":     subject,
		"permission":  permission,
		"status":      "granted",
	})
}

func (h *handlers) revokeWorkspaceProjectPerm(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ws, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	projectKey, err := requireString(req, "project_key")
	if err != nil {
		return errResultErr(err), nil
	}
	subjectSlug, err := requireString(req, "subject_slug")
	if err != nil {
		return errResultErr(err), nil
	}
	isGroup := req.GetBool("is_group", false)
	wpc, err := h.resolveWorkspaceProjectPermsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := wpc.RevokeWorkspaceProjectPerm(ws, projectKey, subjectSlug, isGroup); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{
		"workspace":    ws,
		"project_key":  projectKey,
		"subject_slug": subjectSlug,
		"status":       "revoked",
	})
}
