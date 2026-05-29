package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// ── domain converters ─────────────────────────────────────────────────────────

func toReportDomain(w servergen.RestReport) backend.CodeInsightsReport {
	data := make([]backend.CodeInsightsReportDatum, 0, len(w.Data))
	for _, d := range w.Data {
		data = append(data, backend.CodeInsightsReportDatum{
			Title: d.Title,
			Type:  d.Type,
			Value: d.Value,
		})
	}
	r := backend.CodeInsightsReport{
		Key:        w.Key,
		Title:      w.Title,
		Result:     w.Result,
		ReportType: w.ReportType,
		Details:    w.Details,
		Reporter:   w.Reporter,
		Link:       w.Link,
		LogoURL:    w.LogoURL,
		Data:       data,
	}
	if w.CreatedDate != nil {
		r.CreatedAt = fmt.Sprintf("%d", *w.CreatedDate)
	}
	if w.UpdatedDate != nil {
		r.UpdatedAt = fmt.Sprintf("%d", *w.UpdatedDate)
	}
	return r
}

func toAnnotationDomain(w servergen.RestAnnotation) backend.CodeInsightsAnnotation {
	return backend.CodeInsightsAnnotation{
		ExternalID: w.ExternalID,
		Path:       w.Path,
		Line:       w.Line,
		Message:    w.Message,
		Severity:   w.Severity,
		Type:       w.Type,
		Link:       w.Link,
	}
}

func toGenAnnotation(a backend.CodeInsightsAnnotationInput) servergen.RestAnnotation {
	return servergen.RestAnnotation{
		ExternalID: a.ExternalID,
		Path:       a.Path,
		Line:       a.Line,
		Message:    a.Message,
		Severity:   a.Severity,
		Type:       a.Type,
		Link:       a.Link,
	}
}

func toMergeCheckDomain(w servergen.RestMergeCheck) backend.MergeCheck {
	return backend.MergeCheck{
		Key:         w.Key,
		ReportKey:   w.ReportKey,
		MustPass:    w.MustPass,
		MinSeverity: w.MinSeverity,
	}
}

// ── path helpers ──────────────────────────────────────────────────────────────

func reportBasePath(project, slug, hash string) string {
	return fmt.Sprintf("/projects/%s/repos/%s/commits/%s/reports",
		url.PathEscape(project), url.PathEscape(slug), url.PathEscape(hash))
}

func reportKeyPath(project, slug, hash, key string) string {
	return fmt.Sprintf("%s/%s", reportBasePath(project, slug, hash), url.PathEscape(key))
}

func annotationBasePath(project, slug, hash, key string) string {
	return fmt.Sprintf("%s/annotations", reportKeyPath(project, slug, hash, key))
}

// mergeCheckPath uses /rest/insights/latest per the partly-undocumented API.
// The merge-check key is the check's own identifier (not the report key).
func mergeCheckPath(project, slug, key string) string {
	return fmt.Sprintf("/rest/insights/latest/projects/%s/repos/%s/settings/merge-checks/%s",
		url.PathEscape(project), url.PathEscape(slug), url.PathEscape(key))
}

// ── ListReports ───────────────────────────────────────────────────────────────

// ListReports returns all Code Insights reports attached to a commit.
func (c *Client) ListReports(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
	path := reportBasePath(project, slug, hash)
	return paging.Collect(c.codeInsightsHTTP, path, func(body []byte) ([]backend.CodeInsightsReport, error) {
		var page PagedResponse[servergen.RestReport]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeInsightsReport, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toReportDomain(w))
		}
		return out, nil
	}, 0)
}

// ── GetReport ────────────────────────────────────────────────────────────────

// GetReport fetches a single Code Insights report by key.
func (c *Client) GetReport(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
	var w servergen.RestReport
	if err := c.codeInsightsHTTP.GetJSON(reportKeyPath(project, slug, hash, key), &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return toReportDomain(w), nil
}

// ── SetReport ────────────────────────────────────────────────────────────────

// SetReport creates or replaces a Code Insights report (PUT / upsert).
func (c *Client) SetReport(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
	wdata := make([]servergen.RestReportDatum, 0, len(in.Data))
	for _, d := range in.Data {
		wdata = append(wdata, servergen.RestReportDatum{Title: d.Title, Type: d.Type, Value: d.Value})
	}
	body := servergen.RestReportUpsert{
		Title:      in.Title,
		Result:     in.Result,
		ReportType: in.ReportType,
		Details:    in.Details,
		Reporter:   in.Reporter,
		Link:       in.Link,
		LogoURL:    in.LogoURL,
		Data:       wdata,
	}
	var w servergen.RestReport
	if err := c.codeInsightsHTTP.PutJSON(reportKeyPath(project, slug, hash, key), body, &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return toReportDomain(w), nil
}

// ── DeleteReport ─────────────────────────────────────────────────────────────

// DeleteReport removes a Code Insights report and all its annotations.
func (c *Client) DeleteReport(project, slug, hash, key string) error {
	return c.codeInsightsHTTP.DeleteJSON(reportKeyPath(project, slug, hash, key), nil)
}

// ── ListAnnotations ───────────────────────────────────────────────────────────

// ListAnnotations returns all annotations under a given report key.
func (c *Client) ListAnnotations(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
	path := annotationBasePath(project, slug, hash, key)
	// The annotations endpoint returns {"annotations":[...]} — not the standard
	// paged envelope. We fetch in one shot (BBS docs don't describe pagination
	// here) and return the slice directly.
	var page servergen.RestAnnotationsPage
	if err := c.codeInsightsHTTP.GetJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.CodeInsightsAnnotation, 0, len(page.Annotations))
	for _, w := range page.Annotations {
		out = append(out, toAnnotationDomain(w))
	}
	return out, nil
}

// ── AddAnnotations ────────────────────────────────────────────────────────────

// AddAnnotations bulk-POSTs all annotations in a single request.
func (c *Client) AddAnnotations(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
	wanns := make([]servergen.RestAnnotation, 0, len(in))
	for _, a := range in {
		wanns = append(wanns, toGenAnnotation(a))
	}
	type annotationsBody struct {
		Annotations []servergen.RestAnnotation `json:"annotations"`
	}
	body := annotationsBody{Annotations: wanns}
	return c.codeInsightsHTTP.PostJSON(annotationBasePath(project, slug, hash, key), body, nil)
}

// ── DeleteAnnotations ─────────────────────────────────────────────────────────

// DeleteAnnotations removes all annotations under a given report key.
func (c *Client) DeleteAnnotations(project, slug, hash, key string) error {
	return c.codeInsightsHTTP.DeleteJSON(annotationBasePath(project, slug, hash, key), nil)
}

// ── SetMergeCheck (experimental) ─────────────────────────────────────────────

// SetMergeCheck creates or replaces a merge-check configuration.
// Uses the partly-undocumented /rest/insights/latest/.../merge-check/ path.
func (c *Client) SetMergeCheck(project, slug, key string, in backend.MergeCheckInput) error {
	body := servergen.RestMergeCheck{
		Key:         key,
		ReportKey:   in.ReportKey,
		MustPass:    in.MustPass,
		MinSeverity: in.MinSeverity,
	}
	return c.http.PutJSON(mergeCheckPath(project, slug, key), body, nil)
}

// ── GetMergeCheck (experimental) ─────────────────────────────────────────────

// GetMergeCheck fetches the current merge-check configuration for a key.
func (c *Client) GetMergeCheck(project, slug, key string) (backend.MergeCheck, error) {
	var w servergen.RestMergeCheck
	if err := c.http.GetJSON(mergeCheckPath(project, slug, key), &w); err != nil {
		return backend.MergeCheck{}, err
	}
	return toMergeCheckDomain(w), nil
}

// ── DeleteMergeCheck (experimental) ──────────────────────────────────────────

// DeleteMergeCheck removes a merge-check configuration.
func (c *Client) DeleteMergeCheck(project, slug, key string) error {
	return c.http.DeleteJSON(mergeCheckPath(project, slug, key), nil)
}
