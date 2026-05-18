package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerBranchProtectTools)
}

func registerBranchProtectTools(s *mcpserver.MCPServer, h *handlers) {
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
}
