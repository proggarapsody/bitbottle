package backend

// PipelineTriggerClient is implemented only by Bitbucket Cloud clients.
type PipelineTriggerClient interface {
	TriggerPipeline(ns, slug string, input PipelineTriggerInput) (PipelineTriggerResult, error)
}

// FeaturePipelineTrigger names the pipeline-trigger capability for typed-error reporting.
const FeaturePipelineTrigger Feature = "pipeline_trigger"

// AsPipelineTriggerClient returns the PipelineTriggerClient view of c, or a
// typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does
// not implement the PipelineTrigger capability.
func AsPipelineTriggerClient(c Client, host string) (PipelineTriggerClient, error) {
	return requireFeature[PipelineTriggerClient](c, host, specFor(FeaturePipelineTrigger))
}
