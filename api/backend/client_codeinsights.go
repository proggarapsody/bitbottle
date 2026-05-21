package backend

// CodeInsightsClient exposes Bitbucket Server / Data Center Code Insights
// management. Cloud has no native equivalent — AsCodeInsightsClient returns
// ErrUnsupportedOnHost when called against a Cloud backend.
//
// The merge-check sub-API (SetMergeCheck / GetMergeCheck / DeleteMergeCheck)
// uses a partly undocumented endpoint at /rest/insights/latest/.../merge-check/
// and is marked experimental in CLI help text.
type CodeInsightsClient interface {
	// Report operations
	ListReports(project, slug, hash string) ([]CodeInsightsReport, error)
	GetReport(project, slug, hash, key string) (CodeInsightsReport, error)
	// SetReport upserts (PUT) a report by key.
	SetReport(project, slug, hash, key string, in CodeInsightsReportInput) (CodeInsightsReport, error)
	DeleteReport(project, slug, hash, key string) error

	// Annotation operations
	ListAnnotations(project, slug, hash, key string) ([]CodeInsightsAnnotation, error)
	// AddAnnotations bulk-POSTs all annotations in a single request.
	AddAnnotations(project, slug, hash, key string, in []CodeInsightsAnnotationInput) error
	DeleteAnnotations(project, slug, hash, key string) error

	// Merge-check operations (experimental — partly undocumented API)
	SetMergeCheck(project, slug, key string, in MergeCheckInput) error
	GetMergeCheck(project, slug, key string) (MergeCheck, error)
	DeleteMergeCheck(project, slug, key string) error
}

// FeatureCodeInsights names the Code Insights capability for typed-error
// reporting.
const FeatureCodeInsights Feature = "code-insights"

// AsCodeInsightsClient returns the CodeInsightsClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend that
// doesn't implement Code Insights (currently Bitbucket Cloud).
func AsCodeInsightsClient(c Client, host string) (CodeInsightsClient, error) {
	return requireFeature[CodeInsightsClient](c, host, specFor(FeatureCodeInsights))
}
