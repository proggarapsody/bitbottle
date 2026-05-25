package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudPipelineTestReport is the wire representation of the Cloud test report
// summary for a pipeline step.
type cloudPipelineTestReport struct {
	Status            string  `json:"status"`
	TotalCount        int     `json:"total_count"`
	SuccessCount      int     `json:"success_count"`
	FailedCount       int     `json:"failed_count"`
	ErrorCount        int     `json:"error_count"`
	SkippedCount      int     `json:"skipped_count"`
	DurationInSeconds float64 `json:"duration_in_seconds"`
}

// cloudPipelineTestCase is the wire representation of one test case.
type cloudPipelineTestCase struct {
	TestCaseReason    string  `json:"test_case_reason"`
	Status            string  `json:"status"`
	Name              string  `json:"name"`
	ClassName         string  `json:"class_name"`
	DurationInSeconds float64 `json:"duration_in_seconds"`
	ErrorDetails      string  `json:"error_details"`
	ErrorMessage      string  `json:"error_message"`
}

func testReportPath(ws, slug, pipelineUUID, stepUUID string) string {
	return fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/%s/test_reports",
		url.PathEscape(ws),
		url.PathEscape(slug),
		url.PathEscape(braceUUID(pipelineUUID)),
		url.PathEscape(braceUUID(stepUUID)),
	)
}

// GetPipelineTestReport returns the test report summary for a pipeline step.
// GET /2.0/repositories/{ws}/{slug}/pipelines/{uuid}/steps/{step_uuid}/test_reports
func (c *Client) GetPipelineTestReport(ws, slug, pipelineUUID, stepUUID string) (backend.PipelineTestReport, error) {
	path := testReportPath(ws, slug, pipelineUUID, stepUUID)
	var w cloudPipelineTestReport
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineTestReport{}, err
	}
	return backend.PipelineTestReport{
		Total:      w.TotalCount,
		Passed:     w.SuccessCount,
		Failed:     w.FailedCount + w.ErrorCount,
		Skipped:    w.SkippedCount,
		DurationMS: int64(w.DurationInSeconds * 1000),
	}, nil
}

// ListPipelineTestCases returns the test cases for a pipeline step.
// GET /2.0/repositories/{ws}/{slug}/pipelines/{uuid}/steps/{step_uuid}/test_reports/test_cases
func (c *Client) ListPipelineTestCases(ws, slug, pipelineUUID, stepUUID string, filter backend.TestCaseFilter) ([]backend.PipelineTestCase, error) {
	base := testReportPath(ws, slug, pipelineUUID, stepUUID) + "/test_cases"
	q := url.Values{}
	if filter.Status != "" {
		q.Set("status", filter.Status)
	}
	limit := filter.Limit
	if limit > 0 && limit <= 100 {
		q.Set("pagelen", fmt.Sprintf("%d", limit))
	}
	path := base
	if len(q) > 0 {
		path = base + "?" + q.Encode()
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineTestCase, error) {
		var page cloudPagedResponse[cloudPipelineTestCase]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PipelineTestCase, 0, len(page.Values))
		for _, w := range page.Values {
			failMsg := w.ErrorDetails
			if failMsg == "" {
				failMsg = w.ErrorMessage
			}
			out = append(out, backend.PipelineTestCase{
				Name:           w.Name,
				ClassName:      w.ClassName,
				Status:         w.Status,
				DurationMS:     int64(w.DurationInSeconds * 1000),
				FailureMessage: failMsg,
			})
		}
		return out, nil
	}, limit)
}
