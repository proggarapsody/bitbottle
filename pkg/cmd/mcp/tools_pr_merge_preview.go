package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRMergePreviewTools)
}

func registerPRMergePreviewTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("dry_run_merge_pr",
			mcplib.WithDescription("Preview a pull request merge without merging (dry-run conflict check)"),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
			mcplib.WithString("project",
				mcplib.Description("Project key (Server/DC) or workspace slug (Cloud)"),
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
			mcplib.WithString("strategy",
				mcplib.Description("Merge strategy: ff, squash, merge-commit (optional; ignored on Server dry-run)"),
			),
		),
		h.dryRunMergePR,
	)
}
