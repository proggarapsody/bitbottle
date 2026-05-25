package backend

// PipelineConfigClient is implemented by Cloud backends.
// Bitbucket Server/DC does not have a repository-level pipeline config API.
type PipelineConfigClient interface {
	GetPipelinesConfig(ws, slug string) (PipelineConfig, error)
	UpdatePipelinesConfig(ws, slug string, input PipelineConfig) (PipelineConfig, error)
}

// FeaturePipelineConfig names the pipeline-config capability for typed-error reporting.
const FeaturePipelineConfig Feature = "pipeline_config"

// AsPipelineConfigClient returns the PipelineConfigClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the PipelineConfig capability.
func AsPipelineConfigClient(c Client, host string) (PipelineConfigClient, error) {
	return requireFeature[PipelineConfigClient](c, host, specFor(FeaturePipelineConfig))
}
