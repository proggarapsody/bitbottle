package backend

import "io"

// PipelineClient is implemented only by Bitbucket Cloud clients.
type PipelineClient interface {
	ListPipelines(ns, slug string, limit int) ([]Pipeline, error)
	GetPipeline(ns, slug, uuid string) (Pipeline, error)
	RunPipeline(ns, slug string, in RunPipelineInput) (Pipeline, error)
	StopPipeline(ws, slug, pipelineUUID string) error
	RerunPipeline(ns, slug, sourceUUID string) (Pipeline, error)
	ListPipelineSteps(ns, slug, uuid string) ([]PipelineStep, error)
	GetPipelineStepLog(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error)
	ListPipelineVariables(ns, slug string) ([]PipelineVariable, error)
	SetPipelineVariable(ns, slug string, in PipelineVariableInput) (PipelineVariable, error)
	DeletePipelineVariable(ns, slug, key string) error
}

// FeaturePipelines names the pipelines capability for typed-error reporting.
const FeaturePipelines Feature = "pipelines"

// AsPipelineClient returns the PipelineClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the Pipelines capability.
func AsPipelineClient(c Client, host string) (PipelineClient, error) {
	return requireFeature[PipelineClient](c, host, specFor(FeaturePipelines))
}
