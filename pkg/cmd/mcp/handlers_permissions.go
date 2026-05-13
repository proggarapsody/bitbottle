package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePermissionsClient resolves the backend and type-asserts to
// PermissionsClient, returning a host.unsupported error on Cloud.
func (h *handlers) resolvePermissionsClient(req mcplib.CallToolRequest) (backend.PermissionsClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	pc, err := backend.AsPermissionsClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return pc, hostname, nil
}

// subjectFromReq extracts a PermissionSubject from the tool request.
// Exactly one of "user" or "group" must be provided.
func subjectFromReq(req mcplib.CallToolRequest) (backend.PermissionSubject, error) {
	user := req.GetString("user", "")
	group := req.GetString("group", "")
	if (user == "") == (group == "") {
		return backend.PermissionSubject{}, fmt.Errorf("specify exactly one of user or group")
	}
	if user != "" {
		return backend.PermissionSubject{Kind: "user", Slug: user}, nil
	}
	return backend.PermissionSubject{Kind: "group", Name: group}, nil
}

// ── Project permissions ───────────────────────────────────────────────────────

func (h *handlers) listProjectPermissions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	grants, err := pc.ListProjectPermissions(context.Background(), project)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(grants)
}

func (h *handlers) grantProjectPermission(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	perm, err := requireString(req, "permission")
	if err != nil {
		return errResultErr(err), nil
	}
	subject, err := subjectFromReq(req)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.GrantProjectPermission(context.Background(), project, subject, perm); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "granted", "project": project, "permission": perm})
}

func (h *handlers) revokeProjectPermission(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	subject, err := subjectFromReq(req)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.RevokeProjectPermission(context.Background(), project, subject); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "revoked", "project": project})
}

// ── Repo permissions ──────────────────────────────────────────────────────────

func (h *handlers) listRepoPermissions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	grants, err := pc.ListRepoPermissions(context.Background(), project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(grants)
}

func (h *handlers) grantRepoPermission(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	perm, err := requireString(req, "permission")
	if err != nil {
		return errResultErr(err), nil
	}
	subject, err := subjectFromReq(req)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.GrantRepoPermission(context.Background(), project, slug, subject, perm); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "granted", "project": project, "slug": slug, "permission": perm})
}

func (h *handlers) revokeRepoPermission(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	subject, err := subjectFromReq(req)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, _, err := h.resolvePermissionsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := pc.RevokeRepoPermission(context.Background(), project, slug, subject); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "revoked", "project": project, "slug": slug})
}
