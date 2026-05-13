package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCommitStatusReportTools)
}

func registerCommitStatusReportTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as PROJECT/REPO or WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("report_commit_status",
			mcplib.WithDescription("Post a build status against a commit hash"),
			optHostname,
			reqRepo,
			mcplib.WithString("hash",
				mcplib.Description("Commit hash"),
				mcplib.Required(),
			),
			mcplib.WithString("key",
				mcplib.Description("Build key"),
				mcplib.Required(),
			),
			mcplib.WithString("state",
				mcplib.Description("Build state: SUCCESSFUL, FAILED, INPROGRESS, or STOPPED"),
				mcplib.Required(),
			),
			mcplib.WithString("url",
				mcplib.Description("URL to link with the build status"),
			),
			mcplib.WithString("name",
				mcplib.Description("Display name for the build status"),
			),
			mcplib.WithString("description",
				mcplib.Description("Description for the build status"),
			),
		),
		h.reportCommitStatus,
	)
}
