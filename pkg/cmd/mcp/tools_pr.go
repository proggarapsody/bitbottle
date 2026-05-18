package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRTools)
}

func registerPRTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key or workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	optLimit := mcplib.WithNumber("limit",
		mcplib.Description("Maximum number of results to return"),
	)
	reqID := mcplib.WithNumber("id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_prs",
			mcplib.WithDescription("List pull requests for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("state",
				mcplib.Description("PR state filter: OPEN, MERGED, DECLINED (default: OPEN)"),
			),
			optLimit,
		),
		h.listPRs,
	)

	s.AddTool(
		mcplib.NewTool("get_pr",
			mcplib.WithDescription("Get a single pull request"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
		),
		h.getPR,
	)

	s.AddTool(
		mcplib.NewTool("create_pr",
			mcplib.WithDescription("Create a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("title",
				mcplib.Description("Pull request title"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("Pull request description"),
			),
			mcplib.WithString("from_branch",
				mcplib.Description("Source branch"),
				mcplib.Required(),
			),
			mcplib.WithString("to_branch",
				mcplib.Description("Target branch"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("draft",
				mcplib.Description("Create as draft PR"),
			),
		),
		h.createPR,
	)

	s.AddTool(
		mcplib.NewTool("merge_pr",
			mcplib.WithDescription("Merge a pull request (destructive), or queue it for auto-merge"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithString("strategy",
				mcplib.Description("Merge strategy for immediate merge: merge, squash, rebase (default: server default)"),
			),
			mcplib.WithBoolean("auto",
				mcplib.Description("Queue PR for auto-merge when all checks pass instead of merging immediately"),
			),
			mcplib.WithString("auto_strategy",
				mcplib.Description("Merge strategy for auto-merge: merge, squash, rebase (default: merge)"),
			),
		),
		h.mergePR,
	)

	s.AddTool(
		mcplib.NewTool("approve_pr",
			mcplib.WithDescription("Approve a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
		),
		h.approvePR,
	)

	s.AddTool(
		mcplib.NewTool("get_pr_diff",
			mcplib.WithDescription("Get the unified diff for a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
		),
		h.getPRDiff,
	)

	s.AddTool(
		mcplib.NewTool("update_pr",
			mcplib.WithDescription("Update the title and/or description of a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithString("title",
				mcplib.Description("New pull request title"),
			),
			mcplib.WithString("body",
				mcplib.Description("New pull request description"),
			),
		),
		h.updatePR,
	)

	s.AddTool(
		mcplib.NewTool("decline_pr",
			mcplib.WithDescription("Decline a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.declinePR,
	)

	s.AddTool(
		mcplib.NewTool("reopen_pr",
			mcplib.WithDescription("Reopen a previously declined pull request (Bitbucket Server / DC only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.reopenPR,
	)

	s.AddTool(
		mcplib.NewTool("unapprove_pr",
			mcplib.WithDescription("Remove approval from a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.unapprovePR,
	)

	s.AddTool(
		mcplib.NewTool("ready_pr",
			mcplib.WithDescription("Mark a draft pull request as ready for review"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.readyPR,
	)

	s.AddTool(
		mcplib.NewTool("unready_pull_request",
			mcplib.WithDescription("Convert an open pull request back to draft state"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.unreadyPR,
	)

	s.AddTool(
		mcplib.NewTool("request_review",
			mcplib.WithDescription("Request reviewers on a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithString("reviewers",
				mcplib.Description("Comma-separated list of reviewer usernames/account IDs"),
				mcplib.Required(),
			),
		),
		h.requestReview,
	)
}
