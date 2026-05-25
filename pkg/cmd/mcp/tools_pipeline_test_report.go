package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineTestReportTools)
}

func registerPipelineTestReportTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Workspace or project key"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	reqPipelineUUID := mcplib.WithString("pipeline_uuid",
		mcplib.Description("Pipeline UUID"),
		mcplib.Required(),
	)
	reqStepUUID := mcplib.WithString("step_uuid",
		mcplib.Description("Step UUID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("get_pipeline_test_report",
			mcplib.WithDescription("Get the test report summary for a pipeline step (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqPipelineUUID,
			reqStepUUID,
		),
		h.getPipelineTestReport,
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_test_cases",
			mcplib.WithDescription("List test cases for a pipeline step (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqPipelineUUID,
			reqStepUUID,
			mcplib.WithString("status",
				mcplib.Description("Filter by status: PASSED, FAILED, or SKIPPED"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of test cases (default 50)"),
			),
		),
		h.listPipelineTestCases,
	)
}
