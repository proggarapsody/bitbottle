package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoEditTools)
}

func registerRepoEditTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("edit_repo",
			mcplib.WithDescription("Update mutable repository metadata fields (description, website, language, fork policy, issues/wiki toggles). On Bitbucket Server / Data Center only description is forwarded; Cloud-only fields are accepted but not sent."),
			mcplib.WithString("repo",
				mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
				mcplib.Required(),
			),
			mcplib.WithString("description",
				mcplib.Description("New repository description"),
			),
			mcplib.WithString("website",
				mcplib.Description("Repository website URL (Cloud only)"),
			),
			mcplib.WithString("language",
				mcplib.Description("Repository programming language (Cloud only)"),
			),
			mcplib.WithString("fork_policy",
				mcplib.Description("Fork policy: allow_forks, no_public_forks, or no_forks (Cloud only)"),
			),
			mcplib.WithBoolean("has_issues",
				mcplib.Description("Enable (true) or disable (false) the issue tracker (Cloud only)"),
			),
			mcplib.WithBoolean("has_wiki",
				mcplib.Description("Enable (true) or disable (false) the wiki (Cloud only)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.editRepo,
	)
}
