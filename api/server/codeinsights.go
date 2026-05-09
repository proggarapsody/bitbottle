package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ── wire types ────────────────────────────────────────────────────────────────

// wireReportDatum matches the BBS JSON shape for a single data point.
type wireReportDatum struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// wireReport matches the BBS JSON envelope for a Code Insights report.
type wireReport struct {
	Key         string            `json:"key"`
	Title       string            `json:"title"`
	Result      string            `json:"result"`
	ReportType  string            `json:"reportType"`
	Details     string            `json:"details"`
	Reporter    string            `json:"reporter"`
	Link        string            `json:"link"`
	LogoURL     string            `json:"logoUrl"`
	Data        []wireReportDatum `json:"data"`
	CreatedDate *int64            `json:"createdDate"`
	UpdatedDate *int64            `json:"updatedDate"`
}

func (w wireReport) toDomain() backend.CodeInsightsReport {
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

// wireReportUpsert is the PUT request body for creating/updating a report.
type wireReportUpsert struct {
	Title      string            `json:"title"`
	Result     string            `json:"result,omitempty"`
	ReportType string            `json:"reportType,omitempty"`
	Details    string            `json:"details,omitempty"`
	Reporter   string            `json:"reporter,omitempty"`
	Link       string            `json:"link,omitempty"`
	LogoURL    string            `json:"logoUrl,omitempty"`
	Data       []wireReportDatum `json:"data,omitempty"`
}

// wireAnnotation matches the BBS JSON shape for a single annotation.
type wireAnnotation struct {
	ExternalID string `json:"externalId,omitempty"`
	Path       string `json:"path"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity,omitempty"`
	Type       string `json:"type,omitempty"`
	Link       string `json:"link,omitempty"`
}

func (w wireAnnotation) toDomain() backend.CodeInsightsAnnotation {
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

func toWireAnnotation(a backend.CodeInsightsAnnotationInput) wireAnnotation {
	return wireAnnotation{
		ExternalID: a.ExternalID,
		Path:       a.Path,
		Line:       a.Line,
		Message:    a.Message,
		Severity:   a.Severity,
		Type:       a.Type,
		Link:       a.Link,
	}
}

// wireAnnotationsPage is the paged response for listing annotations.
type wireAnnotationsPage struct {
	Annotations []wireAnnotation `json:"annotations"`
}

// wireMergeCheck matches the BBS JSON shape for a merge-check configuration.
type wireMergeCheck struct {
	Key         string `json:"key"`
	ReportKey   string `json:"reportKey"`
	MustPass    bool   `json:"mustPass"`
	MinSeverity string `json:"minSeverity,omitempty"`
}

func (w wireMergeCheck) toDomain() backend.MergeCheck {
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
		var page PagedResponse[wireReport]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeInsightsReport, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// ── GetReport ────────────────────────────────────────────────────────────────

// GetReport fetches a single Code Insights report by key.
func (c *Client) GetReport(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
	var w wireReport
	if err := c.codeInsightsHTTP.GetJSON(reportKeyPath(project, slug, hash, key), &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return w.toDomain(), nil
}

// ── SetReport ────────────────────────────────────────────────────────────────

// SetReport creates or replaces a Code Insights report (PUT / upsert).
func (c *Client) SetReport(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
	wdata := make([]wireReportDatum, 0, len(in.Data))
	for _, d := range in.Data {
		wdata = append(wdata, wireReportDatum{Title: d.Title, Type: d.Type, Value: d.Value})
	}
	body := wireReportUpsert{
		Title:      in.Title,
		Result:     in.Result,
		ReportType: in.ReportType,
		Details:    in.Details,
		Reporter:   in.Reporter,
		Link:       in.Link,
		LogoURL:    in.LogoURL,
		Data:       wdata,
	}
	var w wireReport
	if err := c.codeInsightsHTTP.PutJSON(reportKeyPath(project, slug, hash, key), body, &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return w.toDomain(), nil
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
	var page wireAnnotationsPage
	if err := c.codeInsightsHTTP.GetJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.CodeInsightsAnnotation, 0, len(page.Annotations))
	for _, w := range page.Annotations {
		out = append(out, w.toDomain())
	}
	return out, nil
}

// ── AddAnnotations ────────────────────────────────────────────────────────────

// AddAnnotations bulk-POSTs all annotations in a single request.
func (c *Client) AddAnnotations(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
	wanns := make([]wireAnnotation, 0, len(in))
	for _, a := range in {
		wanns = append(wanns, toWireAnnotation(a))
	}
	body := map[string]any{"annotations": wanns}
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
	body := wireMergeCheck{
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
	var w wireMergeCheck
	if err := c.http.GetJSON(mergeCheckPath(project, slug, key), &w); err != nil {
		return backend.MergeCheck{}, err
	}
	return w.toDomain(), nil
}

// ── DeleteMergeCheck (experimental) ──────────────────────────────────────────

// DeleteMergeCheck removes a merge-check configuration.
func (c *Client) DeleteMergeCheck(project, slug, key string) error {
	return c.http.DeleteJSON(mergeCheckPath(project, slug, key), nil)
}
