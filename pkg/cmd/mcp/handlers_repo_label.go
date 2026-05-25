package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveRepoLabelClient is the small adapter every repo-label handler shares:
// pick the host, dial the backend, type-assert RepoLabelClient, and gather the
// project/slug args.
func (h *handlers) resolveRepoLabelClient(req mcplib.CallToolRequest) (backend.RepoLabelClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	rl, err := backend.AsRepoLabelClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return rl, project, slug, nil
}

func (h *handlers) listRepoLabels(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	rl, project, slug, err := h.resolveRepoLabelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	labels, err := rl.ListRepoLabels(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(labels)
}

func (h *handlers) createRepoLabel(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	color := req.GetString("color", "")
	rl, project, slug, err := h.resolveRepoLabelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	lbl, err := rl.CreateRepoLabel(project, slug, backend.CreateRepoLabelInput{
		Name:  name,
		Color: color,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(lbl)
}

func (h *handlers) updateRepoLabel(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	name := req.GetString("name", "")
	color := req.GetString("color", "")
	rl, project, slug, err := h.resolveRepoLabelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	lbl, err := rl.UpdateRepoLabel(project, slug, id, backend.UpdateRepoLabelInput{
		Name:  name,
		Color: color,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(lbl)
}

func (h *handlers) deleteRepoLabel(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	rl, project, slug, err := h.resolveRepoLabelClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := rl.DeleteRepoLabel(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "deleted": true})
}
