package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePipelineTestReportClient is the shared preamble for test-report handlers.
func (h *handlers) resolvePipelineTestReportClient(req mcplib.CallToolRequest) (backend.PipelineTestReportClient, string, string, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", "", "", err
	}
	pipelineUUID, err := requireString(req, "pipeline_uuid")
	if err != nil {
		return nil, "", "", "", "", err
	}
	stepUUID, err := requireString(req, "step_uuid")
	if err != nil {
		return nil, "", "", "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", "", "", err
	}
	tc, err := backend.AsPipelineTestReportClient(client, hostname)
	if err != nil {
		return nil, "", "", "", "", err
	}
	return tc, project, slug, pipelineUUID, stepUUID, nil
}

func (h *handlers) getPipelineTestReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	tc, project, slug, pipelineUUID, stepUUID, err := h.resolvePipelineTestReportClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	report, err := tc.GetPipelineTestReport(project, slug, pipelineUUID, stepUUID)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(report)
}

func (h *handlers) listPipelineTestCases(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	tc, project, slug, pipelineUUID, stepUUID, err := h.resolvePipelineTestReportClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	status := req.GetString("status", "")
	limit := req.GetInt("limit", 50)
	filter := backend.TestCaseFilter{
		Status: status,
		Limit:  limit,
	}
	cases, err := tc.ListPipelineTestCases(project, slug, pipelineUUID, stepUUID, filter)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cases)
}
