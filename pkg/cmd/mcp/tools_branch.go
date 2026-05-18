package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerBranchTools)
}

func registerBranchTools(s *mcpserver.MCPServer, h *handlers) {
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
}
