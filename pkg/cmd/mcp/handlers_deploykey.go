package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveDeployKeyClient is the shared preamble for all deploy-key handlers:
// parse hostname + repo, dial backend, type-assert DeployKeyClient.
func (h *handlers) resolveDeployKeyClient(req mcplib.CallToolRequest) (backend.DeployKeyClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return nil, "", "", err
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	dk, err := backend.AsDeployKeyClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return dk, ns, slug, nil
}

func (h *handlers) listDeployKeys(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	dk, ns, slug, err := h.resolveDeployKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	keys, err := dk.ListDeployKeys(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(keys)
}

func (h *handlers) addDeployKey(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	label := req.GetString("label", "")
	rawPermission := req.GetString("permission", "read")
	// Validate: only "read" and "read-write" are accepted.
	if rawPermission != "read" && rawPermission != "read-write" {
		return errResultErr(fmt.Errorf("invalid permission %q: must be read or read-write", rawPermission)), nil
	}
	// Map MCP-facing "read-write" → API wire value "read_write"; "read" passes through.
	var wirePermission string
	switch rawPermission {
	case "read-write":
		wirePermission = "read_write"
	default:
		wirePermission = rawPermission // "read" passes through unchanged
	}
	dk, ns, slug, err := h.resolveDeployKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	added, err := dk.AddDeployKey(ns, slug, backend.DeployKeyInput{Key: key, Label: label, Permission: wirePermission})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(added)
}

func (h *handlers) deleteDeployKey(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	dk, ns, slug, err := h.resolveDeployKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := dk.DeleteDeployKey(ns, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "deleted"})
}
