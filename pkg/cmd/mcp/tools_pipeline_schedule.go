package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineScheduleTools)
}

func registerPipelineScheduleTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqRepo := mcplib.WithString("repo",
		mcplib.Description("Repository as WORKSPACE/REPO"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_schedules",
			mcplib.WithDescription("List pipeline schedules for a repository"),
			optHostname,
			reqRepo,
		),
		h.listPipelineSchedules,
	)

	s.AddTool(
		mcplib.NewTool("create_pipeline_schedule",
			mcplib.WithDescription("Create a pipeline schedule for a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("cron",
				mcplib.Description("Cron expression (e.g. \"0 0 * * *\")"),
				mcplib.Required(),
			),
			mcplib.WithString("branch",
				mcplib.Description("Branch to run the pipeline on"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("enabled",
				mcplib.Description("Whether the schedule is enabled (default true)"),
			),
		),
		h.createPipelineSchedule,
	)

	s.AddTool(
		mcplib.NewTool("delete_pipeline_schedule",
			mcplib.WithDescription("Delete a pipeline schedule from a repository"),
			optHostname,
			reqRepo,
			mcplib.WithString("uuid",
				mcplib.Description("Schedule UUID"),
				mcplib.Required(),
			),
		),
		h.deletePipelineSchedule,
	)
}
