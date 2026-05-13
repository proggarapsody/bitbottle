package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRTaskTools)
}

func registerPRTaskTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_pr_tasks",
			mcplib.WithDescription("List tasks (BLOCKER-severity comments) on a pull request. On Bitbucket Cloud tasks are not supported; all comments are returned without severity filtering."),
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
			mcplib.WithString("state",
				mcplib.Description("Filter by task state: open (default), resolved, all"),
			),
		),
		h.listPRTasks,
	)

	s.AddTool(
		mcplib.NewTool("create_pr_task",
			mcplib.WithDescription("Create a task (BLOCKER comment) on a pull request. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
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
			mcplib.WithString("body",
				mcplib.Description("Task body text"),
				mcplib.Required(),
			),
			mcplib.WithNumber("parent_comment_id",
				mcplib.Description("Anchor task as reply under an existing comment by its ID"),
			),
		),
		h.createPRTask,
	)

	s.AddTool(
		mcplib.NewTool("resolve_pr_task",
			mcplib.WithDescription("Resolve a task on a pull request. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
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
			mcplib.WithNumber("task_id",
				mcplib.Description("Task (comment) ID"),
				mcplib.Required(),
			),
		),
		h.resolvePRTask,
	)

	s.AddTool(
		mcplib.NewTool("reopen_pr_task",
			mcplib.WithDescription("Reopen a resolved task on a pull request. Bitbucket Server / DC only."),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key"),
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
			mcplib.WithNumber("task_id",
				mcplib.Description("Task (comment) ID"),
				mcplib.Required(),
			),
		),
		h.reopenPRTask,
	)
}
