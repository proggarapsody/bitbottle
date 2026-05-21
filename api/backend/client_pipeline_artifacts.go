package backend

import "io"

// PipelineArtifactClient is implemented only by Bitbucket Cloud clients.
type PipelineArtifactClient interface {
	ListPipelineArtifacts(ws, slug, pipelineUUID, stepUUID string, limit int) ([]PipelineArtifact, error)
	DownloadPipelineArtifact(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error
}

// PipelineArtifact represents a single build artifact produced by a pipeline step.
type PipelineArtifact struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	URL       string `json:"url"`
}

// FeaturePipelineArtifacts names the pipeline-artifacts capability for typed-error reporting.
const FeaturePipelineArtifacts Feature = "pipeline-artifacts"

// AsPipelineArtifactClient returns the PipelineArtifactClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host has no
// pipeline-artifacts primitive (Bitbucket Server / Data Center).
func AsPipelineArtifactClient(c Client, host string) (PipelineArtifactClient, error) {
	return requireFeature[PipelineArtifactClient](c, host, specFor(FeaturePipelineArtifacts))
}
