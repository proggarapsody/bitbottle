package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveCloudCIClient resolves the backend and type-asserts to
// CloudCodeInsightsClient, returning a host.unsupported error on Server/DC.
func (h *handlers) resolveCloudCIClient(req mcplib.CallToolRequest) (backend.CloudCodeInsightsClient, string, string, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", "", "", err
	}
	hash := req.GetString("hash", "")
	key := req.GetString("key", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", "", "", err
	}
	ci, err := backend.AsCloudCodeInsightsClient(client, hostname)
	if err != nil {
		return nil, "", "", "", "", err
	}
	return ci, project, slug, hash, key, nil
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (h *handlers) listCloudCIReports(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, _, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	reports, err := ci.ListCodeInsightsReports(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reports)
}

func (h *handlers) getCloudCIReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	r, err := ci.GetCodeInsightsReport(project, slug, hash, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(r)
}

func (h *handlers) putCloudCIReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	title, err := requireString(req, "title")
	if err != nil {
		return errResultErr(err), nil
	}
	result, err := requireString(req, "result")
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.CodeInsightsReportInput{
		Title:      title,
		Result:     strings.ToUpper(result),
		ReportType: strings.ToUpper(req.GetString("report_type", "")),
		Details:    req.GetString("details", ""),
		Reporter:   req.GetString("reporter", ""),
		Link:       req.GetString("link", ""),
		LogoURL:    req.GetString("logo_url", ""),
	}
	r, err := ci.PutCodeInsightsReport(project, slug, hash, key, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(r)
}

func (h *handlers) deleteCloudCIReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	if err := ci.DeleteCodeInsightsReport(project, slug, hash, key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "key": key})
}

// ── Annotations ───────────────────────────────────────────────────────────────

func (h *handlers) listCloudCIAnnotations(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	anns, err := ci.ListCodeInsightsAnnotations(project, slug, hash, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(anns)
}

func (h *handlers) putCloudCIAnnotations(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCloudCIClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	annotationsJSON, err := requireString(req, "annotations_json")
	if err != nil {
		return errResultErr(err), nil
	}
	var anns []backend.CodeInsightsAnnotationInput
	if jerr := json.Unmarshal([]byte(annotationsJSON), &anns); jerr != nil {
		return errResult(fmt.Sprintf("invalid annotations_json: %v", jerr)), nil
	}
	if err := ci.PutCodeInsightsAnnotations(project, slug, hash, key, anns); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"added": len(anns)})
}
