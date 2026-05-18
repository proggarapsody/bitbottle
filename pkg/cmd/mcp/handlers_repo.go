package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listRepos(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	namespace := req.GetString("namespace", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repos, err := client.ListRepos(namespace, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repos)
}

func (h *handlers) getRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	repo, err := client.GetRepo(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) createRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	description := req.GetString("description", "")
	private := req.GetBool("private", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := client.CreateRepo(project, backend.CreateRepoInput{
		Name:        name,
		SCM:         "git",
		Public:      !private,
		Description: description,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) renameRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	newName, err := requireString(req, "new_name")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := client.RenameRepo(project, slug, newName)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) forkRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	into, err := requireString(req, "into")
	if err != nil {
		return errResultErr(err), nil
	}
	name := req.GetString("name", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	forker, err := backend.AsRepoForker(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	fork, err := forker.ForkRepo(project, slug, backend.ForkRepoInput{
		Workspace: into,
		Name:      name,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(fork)
}

func (h *handlers) deleteRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	if err := client.DeleteRepo(project, slug); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}
