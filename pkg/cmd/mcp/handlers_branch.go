package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp/argval"
)

func (h *handlers) listBranches(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

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
	branches, err := client.ListBranches(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(branches)
}

func (h *handlers) createBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	// MCP-12: reject Git-invalid branch names ("/", leading/trailing slashes,
	// "..", control chars, etc.) client-side instead of forwarding to the API.
	name, nameErr := argval.RefName(req.GetArguments(), "name")
	if nameErr != nil {
		return errResultArg(nameErr), nil
	}
	startAt, err := requireString(req, "start_at")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	br, err := client.CreateBranch(project, slug, backend.CreateBranchInput{
		Name:    name,
		StartAt: startAt,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(br)
}

func (h *handlers) deleteBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	branch, err := requireString(req, "branch")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeleteBranch(project, slug, branch); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}
