package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoCoreTools)
}

func registerRepoCoreTools(s *mcpserver.MCPServer, h *handlers) {
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
}
