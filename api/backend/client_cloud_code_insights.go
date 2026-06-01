package backend

// CloudCodeInsightsClient exposes Bitbucket Cloud Code Insights reports
// and annotations. Server/DC uses CodeInsightsClient instead.
type CloudCodeInsightsClient interface {
	ListCodeInsightsReports(project, slug, hash string) ([]CodeInsightsReport, error)
	GetCodeInsightsReport(project, slug, hash, key string) (CodeInsightsReport, error)
	PutCodeInsightsReport(project, slug, hash, key string, in CodeInsightsReportInput) (CodeInsightsReport, error)
	DeleteCodeInsightsReport(project, slug, hash, key string) error
	ListCodeInsightsAnnotations(project, slug, hash, key string) ([]CodeInsightsAnnotation, error)
	PutCodeInsightsAnnotations(project, slug, hash, key string, in []CodeInsightsAnnotationInput) error
}

// FeatureCloudCodeInsights names the Cloud Code Insights capability for
// typed-error reporting.
const FeatureCloudCodeInsights Feature = "cloud-code-insights"

// AsCloudCodeInsightsClient returns the CloudCodeInsightsClient view of c,
// or a typed *DomainError (Kind=ErrUnsupportedOnHost) when called against a
// backend that doesn't implement Cloud Code Insights (currently Bitbucket
// Server / DC).
func AsCloudCodeInsightsClient(c Client, host string) (CloudCodeInsightsClient, error) {
	return requireFeature[CloudCodeInsightsClient](c, host, specFor(FeatureCloudCodeInsights))
}
