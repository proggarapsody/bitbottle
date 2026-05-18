package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineTools)
}

func registerPipelineTools(s *mcpserver.MCPServer, h *handlers) {
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
	reqUUID := mcplib.WithString("uuid",
		mcplib.Description("Pipeline UUID"),
		mcplib.Required(),
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
		mcplib.NewTool("list_pipeline_steps",
			mcplib.WithDescription("List the steps in a pipeline run (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqUUID,
		),
		h.listPipelineSteps,
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_step_log",
			mcplib.WithDescription("Get the plaintext log of a single pipeline step (Bitbucket Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("pipeline_uuid",
				mcplib.Description("Pipeline UUID"),
				mcplib.Required(),
			),
			mcplib.WithString("step_uuid",
				mcplib.Description("Step UUID"),
				mcplib.Required(),
			),
		),
		h.getPipelineStepLog,
	)
}
