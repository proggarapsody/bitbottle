package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerGHPTools)
}

func registerGHPTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("pr_checks",
			mcplib.WithDescription("Show CI/build statuses for a pull request"),
			mcplib.WithString("host",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
		),
		h.prChecks,
	)

	s.AddTool(
		mcplib.NewTool("pr_update_branch",
			mcplib.WithDescription("Sync a PR's source branch with its base branch"),
			mcplib.WithString("host",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithNumber("pr_id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
		),
		h.prUpdateBranch,
	)

	s.AddTool(
		mcplib.NewTool("pr_status",
			mcplib.WithDescription("Show open pull requests authored by or assigned to you"),
			mcplib.WithString("host",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
		),
		h.prStatus,
	)

	s.AddTool(
		mcplib.NewTool("status",
			mcplib.WithDescription("Show your open pull requests across the repository"),
			mcplib.WithString("host",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
		),
		h.ghpStatus,
	)

	s.AddTool(
		mcplib.NewTool("pipeline_watch",
			mcplib.WithDescription("Poll a pipeline until it reaches a terminal state"),
			mcplib.WithString("host",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key or workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("repo",
				mcplib.Description("Repository slug"),
				mcplib.Required(),
			),
			mcplib.WithString("uuid",
				mcplib.Description("Pipeline UUID"),
				mcplib.Required(),
			),
			mcplib.WithNumber("interval",
				mcplib.Description("Polling interval in seconds (default 5)"),
			),
		),
		h.pipelineWatch,
	)
}
