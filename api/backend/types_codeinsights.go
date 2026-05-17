package backend

// CodeInsightsReport is a Bitbucket Server / Data Center Code Insights report
// attached to a specific commit. Reports aggregate quality/security/build
// annotations and roll up to a single PASS/FAIL/PENDING result. This type is
// used for both API responses and the input shape (CodeInsightsReportInput).
//
// Result values: "PASS", "FAIL", "PENDING" (empty = PENDING).
// ReportType values: "TESTING", "COVERAGE", "BUG", "SECURITY", "DUPLICATION",
// "DEPENDENCY" (empty uses TESTING as default on Server).
type CodeInsightsReport struct {
	Key        string                    `json:"key"`
	Title      string                    `json:"title"`
	Result     string                    `json:"result"`
	ReportType string                    `json:"report_type,omitempty"`
	Details    string                    `json:"details,omitempty"`
	Reporter   string                    `json:"reporter,omitempty"`
	Link       string                    `json:"link,omitempty"`
	LogoURL    string                    `json:"logo_url,omitempty"`
	Data       []CodeInsightsReportDatum `json:"data,omitempty"`
	CreatedAt  string                    `json:"created_at,omitempty"`
	UpdatedAt  string                    `json:"updated_at,omitempty"`
}

// CodeInsightsReportDatum is a single key/value data point attached to a
// Code Insights report. Type values: "BOOLEAN", "DATE", "DURATION",
// "LINK", "NUMBER", "PERCENTAGE", "TEXT".
type CodeInsightsReportDatum struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// CodeInsightsReportInput is the upsert payload for SetReport. All fields map
// 1:1 to CodeInsightsReport; the separation exists only so the input shape
// stays explicit in the interface signature.
type CodeInsightsReportInput struct {
	Title      string
	Result     string // "PASS", "FAIL", "PENDING"
	ReportType string
	Details    string
	Reporter   string
	Link       string
	LogoURL    string
	Data       []CodeInsightsReportDatum
}

// CodeInsightsAnnotation is a single file/line annotation posted under a
// Code Insights report. This type serves as both the API response shape and
// the input payload (CodeInsightsAnnotationInput is an alias).
//
// Severity values: "LOW", "MEDIUM", "HIGH", "CRITICAL".
// Type values: "VULNERABILITY", "CODE_SMELL", "BUG".
type CodeInsightsAnnotation struct {
	ExternalID string `json:"external_id,omitempty"`
	Path       string `json:"path"`
	Line       int    `json:"line,omitempty"`
	Message    string `json:"message"`
	Severity   string `json:"severity,omitempty"`
	Type       string `json:"type,omitempty"`
	Link       string `json:"link,omitempty"`
}

// CodeInsightsAnnotationInput is an alias of CodeInsightsAnnotation used on
// the write side of the interface to keep the signature intent clear.
type CodeInsightsAnnotationInput = CodeInsightsAnnotation
