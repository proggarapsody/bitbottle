package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveSSHKeyClient is the shared preamble for all ssh-key handlers:
// parse hostname, dial backend, type-assert SSHKeyClient.
func (h *handlers) resolveSSHKeyClient(req mcplib.CallToolRequest) (backend.SSHKeyClient, string, error) {
	hostname := req.GetString("hostname", "")
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	sk, err := backend.AsSSHKeyClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return sk, hostname, nil
}

func (h *handlers) listSSHKeys(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sk, _, err := h.resolveSSHKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	keys, err := sk.ListSSHKeys()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(keys)
}

func (h *handlers) addSSHKey(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	label := req.GetString("label", "")
	sk, _, err := h.resolveSSHKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	added, err := sk.AddSSHKey(backend.SSHKeyInput{Key: key, Label: label})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(added)
}

func (h *handlers) deleteSSHKey(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	sk, _, err := h.resolveSSHKeyClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := sk.DeleteSSHKey(id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "deleted"})
}
