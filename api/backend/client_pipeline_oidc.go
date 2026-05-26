package backend

// PipelineOIDCClient is implemented only by Bitbucket Cloud clients. It
// exposes the OIDC discovery document and JWKS key-set that Bitbucket Cloud
// Pipelines publishes for workload-identity federation (AWS, GCP, Azure).
type PipelineOIDCClient interface {
	GetPipelineOIDCConfig(workspace string) (PipelineOIDCConfig, error)
	GetPipelineOIDCKeys(workspace string) (PipelineOIDCKeys, error)
}

// FeaturePipelineOIDC names the pipeline OIDC capability for typed-error
// reporting.
const FeaturePipelineOIDC Feature = "pipeline_oidc"

// AsPipelineOIDCClient returns the PipelineOIDCClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC
// backend.
func AsPipelineOIDCClient(c Client, host string) (PipelineOIDCClient, error) {
	return requireFeature[PipelineOIDCClient](c, host, specFor(FeaturePipelineOIDC))
}
