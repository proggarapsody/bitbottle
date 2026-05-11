package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRActivityTools)
}

func registerPRActivityTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("get_pr_activity",
			mcplib.WithDescription("Get activity events for a pull request"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("slug",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of events to return (0 = no limit)"),
			),
		),
		h.getPRActivity,
	)
}

func (h *handlers) getPRActivity(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project := req.GetString("project", "")
	if project == "" {
		return errResult("missing required parameter: project"), nil
	}
	slug := req.GetString("slug", "")
	if slug == "" {
		return errResult("missing required parameter: slug"), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	limit := req.GetInt("limit", 0)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	events, err := client.GetPRActivity(project, slug, prID, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(events)
}
