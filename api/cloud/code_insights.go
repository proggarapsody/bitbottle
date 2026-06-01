package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ── wire types ────────────────────────────────────────────────────────────────

// cloudCIReport is the Cloud Code Insights report wire shape.
type cloudCIReport struct {
	ExternalID string `json:"external_id"`
	Title      string `json:"title"`
	ReportType string `json:"report_type,omitempty"`
	Result     string `json:"result,omitempty"`
	Details    string `json:"details,omitempty"`
	Reporter   string `json:"reporter,omitempty"`
	Link       string `json:"link,omitempty"`
	LogoURL    string `json:"logo_url,omitempty"`
	CreatedOn  string `json:"created_on,omitempty"`
	UpdatedOn  string `json:"updated_on,omitempty"`
}

// cloudCIAnnotation is the Cloud Code Insights annotation wire shape.
type cloudCIAnnotation struct {
	ExternalID string `json:"external_id,omitempty"`
	Path       string `json:"path"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"summary,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Type       string `json:"type,omitempty"`
	Link       string `json:"link,omitempty"`
}

// ── domain converters ─────────────────────────────────────────────────────────

func toCIReportDomain(w cloudCIReport) backend.CodeInsightsReport {
	return backend.CodeInsightsReport{
		Key:        w.ExternalID,
		Title:      w.Title,
		Result:     w.Result,
		ReportType: w.ReportType,
		Details:    w.Details,
		Reporter:   w.Reporter,
		Link:       w.Link,
		LogoURL:    w.LogoURL,
		CreatedAt:  w.CreatedOn,
		UpdatedAt:  w.UpdatedOn,
	}
}

func toCIAnnotationDomain(w cloudCIAnnotation) backend.CodeInsightsAnnotation {
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

func toCIAnnotationWire(a backend.CodeInsightsAnnotationInput) cloudCIAnnotation {
	return cloudCIAnnotation{
		ExternalID: a.ExternalID,
		Path:       a.Path,
		Line:       a.Line,
		Message:    a.Message,
		Severity:   a.Severity,
		Type:       a.Type,
		Link:       a.Link,
	}
}

// ── path helpers ──────────────────────────────────────────────────────────────

func cloudCIReportsBasePath(workspace, slug, hash string) string {
	return fmt.Sprintf("/repositories/%s/%s/commit/%s/reports",
		url.PathEscape(workspace), url.PathEscape(slug), url.PathEscape(hash))
}

func cloudCIReportKeyPath(workspace, slug, hash, key string) string {
	return fmt.Sprintf("%s/%s",
		cloudCIReportsBasePath(workspace, slug, hash), url.PathEscape(key))
}

func cloudCIAnnotationsPath(workspace, slug, hash, key string) string {
	return fmt.Sprintf("%s/annotations", cloudCIReportKeyPath(workspace, slug, hash, key))
}

// ── ListCodeInsightsReports ───────────────────────────────────────────────────

// ListCodeInsightsReports returns all Code Insights reports attached to a
// commit. The workspace argument maps to the Cloud "workspace" slug (same as
// the project field used by the CLI).
func (c *Client) ListCodeInsightsReports(workspace, slug, hash string) ([]backend.CodeInsightsReport, error) {
	path := cloudCIReportsBasePath(workspace, slug, hash)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.CodeInsightsReport, error) {
		var page cloudPagedResponse[cloudCIReport]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeInsightsReport, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCIReportDomain(w))
		}
		return out, nil
	}, 0)
}

// ── GetCodeInsightsReport ─────────────────────────────────────────────────────

// GetCodeInsightsReport fetches a single Code Insights report by key.
func (c *Client) GetCodeInsightsReport(workspace, slug, hash, key string) (backend.CodeInsightsReport, error) {
	var w cloudCIReport
	if err := c.getJSON(cloudCIReportKeyPath(workspace, slug, hash, key), &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return toCIReportDomain(w), nil
}

// ── PutCodeInsightsReport ─────────────────────────────────────────────────────

// PutCodeInsightsReport creates or replaces a Code Insights report (PUT /
// upsert). Only non-zero fields in in are sent.
func (c *Client) PutCodeInsightsReport(workspace, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
	body := cloudCIReport{
		ExternalID: key,
		Title:      in.Title,
		Result:     in.Result,
		ReportType: in.ReportType,
		Details:    in.Details,
		Reporter:   in.Reporter,
		Link:       in.Link,
		LogoURL:    in.LogoURL,
	}
	var w cloudCIReport
	if err := c.putJSON(cloudCIReportKeyPath(workspace, slug, hash, key), body, &w); err != nil {
		return backend.CodeInsightsReport{}, err
	}
	return toCIReportDomain(w), nil
}

// ── DeleteCodeInsightsReport ──────────────────────────────────────────────────

// DeleteCodeInsightsReport removes a Code Insights report and all its
// annotations.
func (c *Client) DeleteCodeInsightsReport(workspace, slug, hash, key string) error {
	return c.delete(cloudCIReportKeyPath(workspace, slug, hash, key))
}

// ── ListCodeInsightsAnnotations ───────────────────────────────────────────────

// ListCodeInsightsAnnotations returns all annotations under a given report.
func (c *Client) ListCodeInsightsAnnotations(workspace, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
	path := cloudCIAnnotationsPath(workspace, slug, hash, key)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.CodeInsightsAnnotation, error) {
		var page cloudPagedResponse[cloudCIAnnotation]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CodeInsightsAnnotation, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCIAnnotationDomain(w))
		}
		return out, nil
	}, 0)
}

// ── PutCodeInsightsAnnotations ────────────────────────────────────────────────

// PutCodeInsightsAnnotations bulk-POSTs annotations to a report in a single
// request. Cloud uses POST with a wrapper object; Server uses {"annotations":[...]}.
func (c *Client) PutCodeInsightsAnnotations(workspace, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
	wanns := make([]cloudCIAnnotation, 0, len(in))
	for _, a := range in {
		wanns = append(wanns, toCIAnnotationWire(a))
	}
	type body struct {
		Annotations []cloudCIAnnotation `json:"annotations"`
	}
	return c.postJSON(cloudCIAnnotationsPath(workspace, slug, hash, key), body{Annotations: wanns}, nil)
}
