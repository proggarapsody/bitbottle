package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPRParticipantTools)
}

func registerPRParticipantTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)
	reqPRID := mcplib.WithNumber("pr_id",
		mcplib.Description("Pull request ID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pr_participants",
			mcplib.WithDescription("List participants in a pull request"),
			optHostname,
			reqRepo,
			reqPRID,
		),
		h.listPRParticipants,
	)

	s.AddTool(
		mcplib.NewTool("update_pr_participant",
			mcplib.WithDescription("Update a pull request participant's approval state. Cloud only."),
			optHostname,
			reqRepo,
			reqPRID,
			mcplib.WithString("account_id",
				mcplib.Required(),
				mcplib.Description("Account ID of the participant"),
			),
			mcplib.WithString("state",
				mcplib.Required(),
				mcplib.Description("New approval state: 'approved', 'changes_requested', or '' (neutral/unapprove)"),
			),
		),
		h.updatePRParticipant,
	)
}
