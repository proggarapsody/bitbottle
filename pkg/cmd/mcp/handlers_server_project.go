package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listServerProjects(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	filter := req.GetString("filter", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsServerProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	projects, err := pc.ListServerProjects(filter, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(projects)
}

func (h *handlers) getServerProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsServerProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	p, err := pc.GetServerProject(key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(p)
}

func (h *handlers) createServerProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	description := req.GetString("description", "")
	public := req.GetBool("public", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsServerProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	p, err := pc.CreateServerProject(backend.CreateServerProjectInput{
		Key:         key,
		Name:        name,
		Description: description,
		Public:      public,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(p)
}

func (h *handlers) updateServerProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsServerProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	args, _ := req.Params.Arguments.(map[string]any)
	in := backend.UpdateServerProjectInput{}
	if _, ok := args["name"]; ok {
		name := req.GetString("name", "")
		in.Name = &name
	}
	if _, ok := args["description"]; ok {
		desc := req.GetString("description", "")
		in.Description = &desc
	}
	if _, ok := args["public"]; ok {
		pub := req.GetBool("public", false)
		in.Public = &pub
	}

	p, err := pc.UpdateServerProject(key, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(p)
}

func (h *handlers) deleteServerProject(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsServerProjectClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if err := pc.DeleteServerProject(key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "key": key})
}
