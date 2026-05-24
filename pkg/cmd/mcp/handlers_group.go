package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listGroups(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	filter := req.GetString("filter", "")
	limit := req.GetInt("limit", 100)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gc, err := backend.AsGroupClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	groups, err := gc.ListGroups(filter, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(groups)
}

func (h *handlers) createGroup(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gc, err := backend.AsGroupClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	g, err := gc.CreateGroup(name)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(g)
}

func (h *handlers) deleteGroup(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gc, err := backend.AsGroupClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := gc.DeleteGroup(name); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "name": name})
}

func (h *handlers) listGroupMembers(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	groupName, err := requireString(req, "group")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 100)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gmc, err := backend.AsGroupMemberClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	members, err := gmc.ListGroupMembers(groupName, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(members)
}

func (h *handlers) addGroupMember(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	groupName, err := requireString(req, "group")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gmc, err := backend.AsGroupMemberClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := gmc.AddGroupMember(groupName, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "added", "group": groupName, "user": user})
}

func (h *handlers) removeGroupMember(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	groupName, err := requireString(req, "group")
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := requireString(req, "user")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	gmc, err := backend.AsGroupMemberClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := gmc.RemoveGroupMember(groupName, user); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "removed", "group": groupName, "user": user})
}
