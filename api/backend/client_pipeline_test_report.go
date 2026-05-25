package backend

// PipelineTestReport is the domain representation of a Bitbucket Cloud
// pipeline test report summary for a step.
type PipelineTestReport struct {
	Total      int   `json:"total"`
	Passed     int   `json:"passed"`
	Failed     int   `json:"failed"`
	Skipped    int   `json:"skipped"`
	DurationMS int64 `json:"duration_ms"`
}

// PipelineTestCase is the domain representation of one test case in a
// Bitbucket Cloud pipeline step test report.
type PipelineTestCase struct {
	Name           string `json:"name"`
	ClassName      string `json:"class_name"`
	Status         string `json:"status"` // PASSED | FAILED | SKIPPED
	DurationMS     int64  `json:"duration_ms"`
	FailureMessage string `json:"failure_message,omitempty"`
}

// TestCaseFilter controls which test cases are returned by ListPipelineTestCases.
type TestCaseFilter struct {
	Status string // PASSED | FAILED | SKIPPED | "" (all)
	Limit  int
}

// PipelineTestReportClient is implemented by Cloud backends.
// Bitbucket Server/DC does not have a pipeline test-report API.
type PipelineTestReportClient interface {
	GetPipelineTestReport(ws, slug, pipelineUUID, stepUUID string) (PipelineTestReport, error)
	ListPipelineTestCases(ws, slug, pipelineUUID, stepUUID string, filter TestCaseFilter) ([]PipelineTestCase, error)
}

// FeaturePipelineTestReports names the pipeline-test-reports capability.
const FeaturePipelineTestReports Feature = "pipeline_test_reports"

// AsPipelineTestReportClient returns the PipelineTestReportClient view of c,
// or a typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend does not
// implement the capability.
func AsPipelineTestReportClient(c Client, host string) (PipelineTestReportClient, error) {
	return requireFeature[PipelineTestReportClient](c, host, specFor(FeaturePipelineTestReports))
}
