package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoPRSettingsTools)
}

func registerRepoPRSettingsTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("get_repo_pr_settings",
			mcplib.WithDescription("Show pull request gate settings for a repository (required approvers, merge strategy, etc.). Server / Data Center only — returns an error on Cloud."),
			mcplib.WithString("project",
				mcplib.Required(),
				mcplib.Description("Project key (Server) or workspace slug (Cloud)"),
			),
			mcplib.WithString("repo",
				mcplib.Required(),
				mcplib.Description("Repository slug"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.getRepoPRSettings,
	)
	s.AddTool(
		mcplib.NewTool("set_repo_pr_settings",
			mcplib.WithDescription("Update pull request gate settings for a repository (required approvers, merge strategy, allowed strategies, etc.). Server / Data Center only — returns an error on Cloud."),
			mcplib.WithString("project",
				mcplib.Required(),
				mcplib.Description("Project key (Server) or workspace slug (Cloud)"),
			),
			mcplib.WithString("repo",
				mcplib.Required(),
				mcplib.Description("Repository slug"),
			),
			mcplib.WithNumber("required_approvers",
				mcplib.Description("Minimum number of approvals required"),
			),
			mcplib.WithBoolean("required_all_approvers",
				mcplib.Description("Require all reviewers to approve"),
			),
			mcplib.WithBoolean("required_all_tasks_complete",
				mcplib.Description("Require all tasks to be resolved before merge"),
			),
			mcplib.WithNumber("required_successful_builds",
				mcplib.Description("Minimum number of successful builds required"),
			),
			mcplib.WithString("merge_strategy",
				mcplib.Description("Default merge strategy (e.g. no-ff, squash, ff, ff-only, rebase)"),
			),
			mcplib.WithArray("allowed_strategies",
				mcplib.Description("List of allowed merge strategies"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.setRepoPRSettings,
	)
}
