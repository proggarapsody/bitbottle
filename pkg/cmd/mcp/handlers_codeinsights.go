package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveCodeInsightsClient resolves the backend and type-asserts to
// CodeInsightsClient, returning a host.unsupported error on Cloud.
func (h *handlers) resolveCodeInsightsClient(req mcplib.CallToolRequest) (backend.CodeInsightsClient, string, string, string, string, error) {
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
	ci, err := backend.AsCodeInsightsClient(client, hostname)
	if err != nil {
		return nil, "", "", "", "", err
	}
	return ci, project, slug, hash, key, nil
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (h *handlers) listCodeInsightsReports(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, _, err := h.resolveCodeInsightsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	reports, err := ci.ListReports(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reports)
}

func (h *handlers) getCodeInsightsReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	r, err := ci.GetReport(project, slug, hash, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(r)
}

func (h *handlers) setCodeInsightsReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
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
	r, err := ci.SetReport(project, slug, hash, key, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(r)
}

func (h *handlers) deleteCodeInsightsReport(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	if err := ci.DeleteReport(project, slug, hash, key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "key": key})
}

// ── Annotations ───────────────────────────────────────────────────────────────

func (h *handlers) listCodeInsightsAnnotations(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	anns, err := ci.ListAnnotations(project, slug, hash, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(anns)
}

func (h *handlers) addCodeInsightsAnnotations(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
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
	if err := ci.AddAnnotations(project, slug, hash, key, anns); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"added": len(anns)})
}

func (h *handlers) deleteCodeInsightsAnnotations(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ci, project, slug, hash, key, err := h.resolveCodeInsightsClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if hash == "" {
		return errResult("missing required parameter: hash"), nil
	}
	if key == "" {
		return errResult("missing required parameter: key"), nil
	}
	if err := ci.DeleteAnnotations(project, slug, hash, key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted"})
}

// ── Merge checks ─────────────────────────────────────────────────────────────

func (h *handlers) setMergeCheck(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	reportKey, err := requireString(req, "report_key")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ci, err := backend.AsCodeInsightsClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.MergeCheckInput{
		Key:         key,
		ReportKey:   reportKey,
		MustPass:    req.GetBool("must_pass", false),
		MinSeverity: strings.ToUpper(req.GetString("min_severity", "")),
	}
	if err := ci.SetMergeCheck(project, slug, key, in); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "ok", "key": key})
}

func (h *handlers) getMergeCheck(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ci, err := backend.AsCodeInsightsClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	mc, err := ci.GetMergeCheck(project, slug, key)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(mc)
}

func (h *handlers) deleteMergeCheck(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ci, err := backend.AsCodeInsightsClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ci.DeleteMergeCheck(project, slug, key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"status": "deleted", "key": key})
}
