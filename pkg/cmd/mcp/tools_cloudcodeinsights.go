package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerCloudCodeInsightsTools)
}

// registerCloudCodeInsightsTools wires the Cloud Code Insights tools onto the
// MCP server. All tools target Bitbucket Cloud; calls against Server/DC return
// host.unsupported via AsCloudCodeInsightsClient.
func registerCloudCodeInsightsTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqWorkspace := mcplib.WithString("project",
		mcplib.Description("Workspace slug"),
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
		mcplib.Description("Report key (external_id)"),
		mcplib.Required(),
	)

	// ── Reports ──────────────────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_cloud_ci_reports",
			mcplib.WithDescription("List Cloud Code Insights reports attached to a commit (Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash,
		),
		h.listCloudCIReports,
	)

	s.AddTool(
		mcplib.NewTool("get_cloud_ci_report",
			mcplib.WithDescription("Get a single Cloud Code Insights report by key (Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash, reqKey,
		),
		h.getCloudCIReport,
	)

	s.AddTool(
		mcplib.NewTool("put_cloud_ci_report",
			mcplib.WithDescription("Create or update (upsert) a Cloud Code Insights report (Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash, reqKey,
			mcplib.WithString("title",
				mcplib.Description("Report title"),
				mcplib.Required(),
			),
			mcplib.WithString("result",
				mcplib.Description("Report result: PASSED, FAILED, or PENDING"),
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
		h.putCloudCIReport,
	)

	s.AddTool(
		mcplib.NewTool("delete_cloud_ci_report",
			mcplib.WithDescription("Delete a Cloud Code Insights report and all its annotations (destructive; Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash, reqKey,
		),
		h.deleteCloudCIReport,
	)

	// ── Annotations ──────────────────────────────────────────────────────────

	s.AddTool(
		mcplib.NewTool("list_cloud_ci_annotations",
			mcplib.WithDescription("List annotations for a Cloud Code Insights report (Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash, reqKey,
		),
		h.listCloudCIAnnotations,
	)

	s.AddTool(
		mcplib.NewTool("put_cloud_ci_annotations",
			mcplib.WithDescription("Bulk-add annotations to a Cloud Code Insights report (Bitbucket Cloud only)"),
			optHostname, reqWorkspace, reqSlug, reqHash, reqKey,
			mcplib.WithString("annotations_json",
				mcplib.Description("JSON array of annotation objects with fields: path, line, summary, severity, type, external_id, link"),
				mcplib.Required(),
			),
		),
		h.putCloudCIAnnotations,
	)
}
