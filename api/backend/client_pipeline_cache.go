package backend

import "fmt"

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
	pc, ok := c.(PipelineCacheClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePipelineCache),
			Message: fmt.Sprintf("pipeline caches are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return pc, nil
}
