package backend

// PipelineCacheClient is implemented only by Bitbucket Cloud clients.
type PipelineCacheClient interface {
	ListPipelineCaches(ns, slug string) ([]PipelineCache, error)
	DeletePipelineCache(ns, slug, uuid string) error
}

// FeaturePipelineCache names the pipeline-cache capability for typed-error reporting.
const FeaturePipelineCache Feature = "pipeline-cache"

// AsPipelineCacheClient returns the PipelineCacheClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host has no
// pipeline-cache primitive (Bitbucket Server / Data Center).
func AsPipelineCacheClient(c Client, host string) (PipelineCacheClient, error) {
	return requireFeature[PipelineCacheClient](c, host, specFor(FeaturePipelineCache))
}
