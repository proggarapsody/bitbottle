package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerBranchRuleTools)
}

func registerBranchRuleTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_branch_rules",
			mcplib.WithDescription("List branch restriction rules for a repository"),
			optHostname,
			reqRepo,
		),
		h.listBranchRules,
	)

	s.AddTool(
		mcplib.NewTool("add_branch_rule",
			mcplib.WithDescription("Add a branch restriction rule to a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("kind",
				mcplib.Description("Branch restriction kind (e.g. push, require_approvals_to_merge)"),
				mcplib.Required(),
			),
			mcplib.WithString("pattern",
				mcplib.Description("Branch pattern to restrict (e.g. main, feature/*)"),
				mcplib.Required(),
			),
			mcplib.WithNumber("value",
				mcplib.Description("Numeric value for the rule (e.g. required approvals count)"),
			),
		),
		h.addBranchRule,
	)

	s.AddTool(
		mcplib.NewTool("delete_branch_rule",
			mcplib.WithDescription("Delete a branch restriction rule from a repository"),
			optHostname,
			reqRepo,
			mcplib.WithNumber("id",
				mcplib.Description("Branch rule ID"),
				mcplib.Required(),
			),
		),
		h.deleteBranchRule,
	)

	s.AddTool(
		mcplib.NewTool("update_branch_rule",
			mcplib.WithDescription("Update a branch restriction rule in a repository"),
			optHostname,
			reqRepo,
			mcplib.WithNumber("id",
				mcplib.Description("Branch rule ID"),
				mcplib.Required(),
			),
			mcplib.WithString("pattern",
				mcplib.Description("New branch pattern (e.g. main, feature/*)"),
			),
			mcplib.WithString("users",
				mcplib.Description("Comma-separated user slugs (replaces existing)"),
			),
			mcplib.WithString("groups",
				mcplib.Description("Comma-separated group slugs (replaces existing)"),
			),
			mcplib.WithNumber("value",
				mcplib.Description("Numeric value for the rule (e.g. required approvers count)"),
			),
		),
		h.updateBranchRule,
	)
}
