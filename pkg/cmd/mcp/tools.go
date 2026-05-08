package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func newMCPServer(f *factory.Factory) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("bitbottle", "0.1.0",
		mcpserver.WithToolCapabilities(false),
	)
	h := newHandlers(f)
	registerTools(s, h)
	return s
}

func registerTools(s *mcpserver.MCPServer, h *handlers) {
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

	s.AddTool(
		mcplib.NewTool("list_hosts",
			mcplib.WithDescription("List all configured Bitbucket hosts"),
		),
		h.listHosts,
	)

	s.AddTool(
		mcplib.NewTool("list_repos",
			mcplib.WithDescription("List repositories on a Bitbucket host"),
			optHostname,
			mcplib.WithString("namespace",
				mcplib.Description("Workspace slug (Bitbucket Cloud) or leave empty for Server"),
			),
			optLimit,
		),
		h.listRepos,
	)

	s.AddTool(
		mcplib.NewTool("get_repo",
			mcplib.WithDescription("Get a single repository"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.getRepo,
	)

	s.AddTool(
		mcplib.NewTool("create_repo",
			mcplib.WithDescription("Create a new repository"),
			optHostname,
			reqProject,
			mcplib.WithString("name",
				mcplib.Description("Repository name"),
				mcplib.Required(),
			),
			mcplib.WithString("description",
				mcplib.Description("Repository description"),
			),
			mcplib.WithBoolean("private",
				mcplib.Description("Whether the repository is private (default: false)"),
			),
		),
		h.createRepo,
	)

	s.AddTool(
		mcplib.NewTool("delete_repo",
			mcplib.WithDescription("Delete a repository (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.deleteRepo,
	)

	s.AddTool(
		mcplib.NewTool("rename_repo",
			mcplib.WithDescription("Rename a repository (changes name and slug; both backends)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("new_name",
				mcplib.Description("New repository name"),
				mcplib.Required(),
			),
		),
		h.renameRepo,
	)

	s.AddTool(
		mcplib.NewTool("fork_repo",
			mcplib.WithDescription("Fork a Bitbucket Cloud repository into a destination workspace (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("into",
				mcplib.Description("Destination workspace slug"),
				mcplib.Required(),
			),
			mcplib.WithString("name",
				mcplib.Description("Override the fork's name (defaults to source name)"),
			),
		),
		h.forkRepo,
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
			mcplib.WithDescription("Merge a pull request (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Pull request ID"),
				mcplib.Required(),
			),
			mcplib.WithString("strategy",
				mcplib.Description("Merge strategy: merge, squash, rebase (default: server default)"),
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
		mcplib.NewTool("delete_branch",
			mcplib.WithDescription("Delete a branch (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("branch",
				mcplib.Description("Branch name to delete"),
				mcplib.Required(),
			),
		),
		h.deleteBranch,
	)

	s.AddTool(
		mcplib.NewTool("get_current_user",
			mcplib.WithDescription("Get the currently authenticated user"),
			optHostname,
		),
		h.getCurrentUser,
	)

	s.AddTool(
		mcplib.NewTool("list_branches",
			mcplib.WithDescription("List branches for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			optLimit,
		),
		h.listBranches,
	)

	s.AddTool(
		mcplib.NewTool("create_branch",
			mcplib.WithDescription("Create a new branch in a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Name for the new branch"),
				mcplib.Required(),
			),
			mcplib.WithString("start_at",
				mcplib.Description("Branch name or commit hash to start the new branch from"),
				mcplib.Required(),
			),
		),
		h.createBranch,
	)

	s.AddTool(
		mcplib.NewTool("list_pipelines",
			mcplib.WithDescription("List pipelines for a repository (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			optLimit,
		),
		h.listPipelines,
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline",
			mcplib.WithDescription("Get a single pipeline by UUID (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("uuid",
				mcplib.Description("Pipeline UUID"),
				mcplib.Required(),
			),
		),
		h.getPipeline,
	)

	s.AddTool(
		mcplib.NewTool("run_pipeline",
			mcplib.WithDescription("Trigger a pipeline on a branch (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("branch",
				mcplib.Description("Branch to run the pipeline on"),
				mcplib.Required(),
			),
		),
		h.runPipeline,
	)

	s.AddTool(
		mcplib.NewTool("list_tags",
			mcplib.WithDescription("List tags for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			optLimit,
		),
		h.listTags,
	)

	s.AddTool(
		mcplib.NewTool("create_tag",
			mcplib.WithDescription("Create a tag in a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Tag name"),
				mcplib.Required(),
			),
			mcplib.WithString("start_at",
				mcplib.Description("Branch name or commit hash to tag"),
				mcplib.Required(),
			),
			mcplib.WithString("message",
				mcplib.Description("Tag message (creates annotated tag when non-empty)"),
			),
		),
		h.createTag,
	)

	s.AddTool(
		mcplib.NewTool("delete_tag",
			mcplib.WithDescription("Delete a tag (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Tag name to delete"),
				mcplib.Required(),
			),
		),
		h.deleteTag,
	)

	reqID := mcplib.WithNumber("id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
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

	s.AddTool(
		mcplib.NewTool("list_commits",
			mcplib.WithDescription("List commits for a repository"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("branch",
				mcplib.Description("Branch to list commits from (default: main)"),
			),
			optLimit,
		),
		h.listCommits,
	)

	s.AddTool(
		mcplib.NewTool("get_commit",
			mcplib.WithDescription("Get a single commit by hash"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
		),
		h.getCommit,
	)

	s.AddTool(
		mcplib.NewTool("list_pr_comments",
			mcplib.WithDescription("List general (top-level) comments on a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
		),
		h.listPRComments,
	)

	s.AddTool(
		mcplib.NewTool("add_pr_comment",
			mcplib.WithDescription("Add a general comment to a pull request"),
			optHostname,
			reqProject,
			reqSlug,
			reqID,
			mcplib.WithString("body",
				mcplib.Description("Comment body"),
				mcplib.Required(),
			),
		),
		h.addPRComment,
	)

	s.AddTool(
		mcplib.NewTool("list_commit_statuses",
			mcplib.WithDescription("List build / CI statuses reported against a commit hash"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
		),
		h.listCommitStatuses,
	)

	reqUUID := mcplib.WithString("uuid",
		mcplib.Description("Pipeline UUID"),
		mcplib.Required(),
	)
	reqKey := mcplib.WithString("key",
		mcplib.Description("Pipeline variable key"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_steps",
			mcplib.WithDescription("List the steps in a pipeline run (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqUUID,
		),
		h.listPipelineSteps,
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_step_log",
			mcplib.WithDescription("Get the plaintext log of a single pipeline step (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("pipeline_uuid",
				mcplib.Description("Pipeline UUID"),
				mcplib.Required(),
			),
			mcplib.WithString("step_uuid",
				mcplib.Description("Step UUID"),
				mcplib.Required(),
			),
		),
		h.getPipelineStepLog,
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_variables",
			mcplib.WithDescription("List repository-level pipeline variables (Bitbucket Cloud only). Secured variable values are not returned."),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listPipelineVariables,
	)

	s.AddTool(
		mcplib.NewTool("set_pipeline_variable",
			mcplib.WithDescription("Create or update a repository-level pipeline variable, upsert by key (destructive write; Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqKey,
			mcplib.WithString("value",
				mcplib.Description("Variable value"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("secured",
				mcplib.Description("Mark as secured (value redacted on read)"),
			),
		),
		h.setPipelineVariable,
	)

	s.AddTool(
		mcplib.NewTool("delete_pipeline_variable",
			mcplib.WithDescription("Delete a repository-level pipeline variable by key (destructive; Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqKey,
		),
		h.deletePipelineVariable,
	)

	s.AddTool(
		mcplib.NewTool("list_webhooks",
			mcplib.WithDescription("List repository webhooks"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listWebhooks,
	)

	s.AddTool(
		mcplib.NewTool("get_webhook",
			mcplib.WithDescription("Get a single webhook by ID"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("id",
				mcplib.Description("Webhook ID"),
				mcplib.Required(),
			),
		),
		h.getWebhook,
	)

	s.AddTool(
		mcplib.NewTool("create_webhook",
			mcplib.WithDescription("Create a repository webhook"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("url",
				mcplib.Description("Webhook delivery URL"),
				mcplib.Required(),
			),
			mcplib.WithArray("events",
				mcplib.Description("Event keys the webhook subscribes to"),
				mcplib.WithStringItems(),
				mcplib.Required(),
			),
			mcplib.WithBoolean("active",
				mcplib.Description("Whether the webhook is active on creation (defaults to true)"),
			),
			mcplib.WithString("secret",
				mcplib.Description("Shared secret for HMAC signing of delivery payloads (optional)"),
			),
		),
		h.createWebhook,
	)

	s.AddTool(
		mcplib.NewTool("delete_webhook",
			mcplib.WithDescription("Delete a webhook by ID (destructive)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("id",
				mcplib.Description("Webhook ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteWebhook,
	)

	s.AddTool(
		mcplib.NewTool("list_workspaces",
			mcplib.WithDescription("List Bitbucket Cloud workspaces the authenticated user belongs to (Cloud only)"),
			optHostname,
			optLimit,
		),
		h.listWorkspaces,
	)

	s.AddTool(
		mcplib.NewTool("list_projects",
			mcplib.WithDescription("List projects within a Bitbucket Cloud workspace (Cloud only)"),
			optHostname,
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug"),
				mcplib.Required(),
			),
			optLimit,
		),
		h.listProjects,
	)

	s.AddTool(
		mcplib.NewTool("list_issues",
			mcplib.WithDescription("List issues in a Bitbucket Cloud repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("state",
				mcplib.Description("State filter (open, new, on hold, resolved, duplicate, invalid, wontfix, closed); empty = all"),
			),
			optLimit,
		),
		h.listIssues,
	)

	s.AddTool(
		mcplib.NewTool("get_issue",
			mcplib.WithDescription("Get a single issue by ID (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Issue ID"),
				mcplib.Required(),
			),
		),
		h.getIssue,
	)

	s.AddTool(
		mcplib.NewTool("create_issue",
			mcplib.WithDescription("Create a new issue (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("title",
				mcplib.Description("Issue title"),
				mcplib.Required(),
			),
			mcplib.WithString("body",
				mcplib.Description("Issue body (markdown)"),
			),
			mcplib.WithString("kind",
				mcplib.Description("bug, enhancement, proposal, task"),
			),
			mcplib.WithString("priority",
				mcplib.Description("trivial, minor, major, critical, blocker"),
			),
		),
		h.createIssue,
	)

	s.AddTool(
		mcplib.NewTool("close_issue",
			mcplib.WithDescription("Close an issue by setting its state to \"closed\" (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Issue ID"),
				mcplib.Required(),
			),
		),
		h.closeIssue,
	)

	s.AddTool(
		mcplib.NewTool("list_branch_protections",
			mcplib.WithDescription("List branch restrictions for a repository (Bitbucket Server / DC only)"),
			optHostname,
			reqProject,
			reqSlug,
			optLimit,
		),
		h.listBranchProtections,
	)

	s.AddTool(
		mcplib.NewTool("create_branch_protection",
			mcplib.WithDescription("Create a branch restriction (Bitbucket Server / DC only). Provide exactly one of branch or pattern."),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("type",
				mcplib.Description("Restriction type: read-only, no-deletes, fast-forward-only, or pull-request-only"),
				mcplib.Required(),
			),
			mcplib.WithString("branch",
				mcplib.Description("Single branch name to restrict (mutually exclusive with pattern)"),
			),
			mcplib.WithString("pattern",
				mcplib.Description("Glob pattern of branches to restrict, e.g. \"release/*\" (mutually exclusive with branch)"),
			),
			mcplib.WithArray("users",
				mcplib.Description("User slugs exempted from the restriction"),
				mcplib.WithStringItems(),
			),
			mcplib.WithArray("groups",
				mcplib.Description("Group slugs exempted from the restriction"),
				mcplib.WithStringItems(),
			),
		),
		h.createBranchProtection,
	)

	s.AddTool(
		mcplib.NewTool("delete_branch_protection",
			mcplib.WithDescription("Delete a branch restriction by numeric ID (destructive; Bitbucket Server / DC only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("id",
				mcplib.Description("Restriction ID to delete"),
				mcplib.Required(),
			),
		),
		h.deleteBranchProtection,
	)

	s.AddTool(
		mcplib.NewTool("search_code",
			mcplib.WithDescription("Search code across a Bitbucket Cloud workspace (Cloud only). Bitbucket's query language is passed through verbatim — operators like 'path:', 'lang:', and exact-phrase quoting work as documented."),
			optHostname,
			mcplib.WithString("workspace",
				mcplib.Description("Workspace slug to scope the search to"),
				mcplib.Required(),
			),
			mcplib.WithString("query",
				mcplib.Description("Bitbucket Cloud code-search query string"),
				mcplib.Required(),
			),
			optLimit,
		),
		h.searchCode,
	)
}
