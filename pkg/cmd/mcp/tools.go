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
}
