package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCodeInsightsTools)
}

// registerCodeInsightsTools wires the Code Insights tools onto the MCP
// server. All tools are Bitbucket Server / DC only — calls against Cloud
// return host.unsupported via AsCodeInsightsClient.
func registerCodeInsightsTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	reqHash := mcplib.WithString("hash",
		mcplib.Description("Commit hash"),
		mcplib.Required(),
	)
	reqKey := mcplib.WithString("key",
		mcplib.Description("Report or check key"),
		mcplib.Required(),
	)

	// ── Reports ──────────────────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_code_insights_reports",
			mcplib.WithDescription("List Code Insights reports attached to a commit (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash,
		),
		h.listCodeInsightsReports,
	)

	s.AddTool(
		mcplib.NewTool("get_code_insights_report",
			mcplib.WithDescription("Get a single Code Insights report by key (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
		),
		h.getCodeInsightsReport,
	)

	s.AddTool(
		mcplib.NewTool("set_code_insights_report",
			mcplib.WithDescription("Create or update (upsert) a Code Insights report (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
			mcplib.WithString("title",
				mcplib.Description("Report title"),
				mcplib.Required(),
			),
			mcplib.WithString("result",
				mcplib.Description("Report result: PASS, FAIL, or PENDING"),
				mcplib.Required(),
			),
			mcplib.WithString("report_type",
				mcplib.Description("Report type: TESTING, COVERAGE, BUG, SECURITY, DUPLICATION, DEPENDENCY"),
			),
			mcplib.WithString("details",
				mcplib.Description("Human-readable details"),
			),
			mcplib.WithString("reporter",
				mcplib.Description("Name of the tool or reporter"),
			),
			mcplib.WithString("link",
				mcplib.Description("URL linking to the full report"),
			),
			mcplib.WithString("logo_url",
				mcplib.Description("URL of the tool logo"),
			),
		),
		h.setCodeInsightsReport,
	)

	s.AddTool(
		mcplib.NewTool("delete_code_insights_report",
			mcplib.WithDescription("Delete a Code Insights report and all its annotations (destructive; Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
		),
		h.deleteCodeInsightsReport,
	)

	// ── Annotations ──────────────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_code_insights_annotations",
			mcplib.WithDescription("List annotations for a Code Insights report (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
		),
		h.listCodeInsightsAnnotations,
	)

	s.AddTool(
		mcplib.NewTool("add_code_insights_annotations",
			mcplib.WithDescription("Bulk-add annotations to a Code Insights report in a single request (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
			mcplib.WithString("annotations_json",
				mcplib.Description("JSON array of annotation objects with fields: path, line, message, severity, type, external_id, link"),
				mcplib.Required(),
			),
		),
		h.addCodeInsightsAnnotations,
	)

	s.AddTool(
		mcplib.NewTool("delete_code_insights_annotations",
			mcplib.WithDescription("Delete all annotations for a Code Insights report (destructive; Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHash, reqKey,
		),
		h.deleteCodeInsightsAnnotations,
	)

	// ── Merge checks (experimental) ──────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("set_merge_check",
			mcplib.WithDescription("Create or update a merge-check configuration. EXPERIMENTAL: uses a partly undocumented Bitbucket Server API."),
			optHostname, reqProject, reqSlug, reqKey,
			mcplib.WithString("report_key",
				mcplib.Description("Code Insights report key this check applies to"),
				mcplib.Required(),
			),
			mcplib.WithBoolean("must_pass",
				mcplib.Description("Block merge when the report does not pass"),
			),
			mcplib.WithString("min_severity",
				mcplib.Description("Minimum annotation severity to block merge: LOW, MEDIUM, HIGH, CRITICAL"),
			),
		),
		h.setMergeCheck,
	)

	s.AddTool(
		mcplib.NewTool("get_merge_check",
			mcplib.WithDescription("Get the current merge-check configuration. EXPERIMENTAL: uses a partly undocumented Bitbucket Server API."),
			optHostname, reqProject, reqSlug, reqKey,
		),
		h.getMergeCheck,
	)

	s.AddTool(
		mcplib.NewTool("delete_merge_check",
			mcplib.WithDescription("Delete a merge-check configuration (destructive). EXPERIMENTAL: uses a partly undocumented Bitbucket Server API."),
			optHostname, reqProject, reqSlug, reqKey,
		),
		h.deleteMergeCheck,
	)
}
