package backend

// PipelineScheduleClient is implemented only by Bitbucket Cloud clients.
type PipelineScheduleClient interface {
	ListPipelineSchedules(ns, slug string) ([]PipelineSchedule, error)
	CreatePipelineSchedule(ns, slug string, input PipelineScheduleInput) (PipelineSchedule, error)
	DeletePipelineSchedule(ns, slug, uuid string) error
}

// FeaturePipelineSchedules names the pipeline-schedules capability for typed-error reporting.
const FeaturePipelineSchedules Feature = "pipeline_schedules"

// AsPipelineScheduleClient returns the PipelineScheduleClient view of c, or a
// typed *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does
// not implement the PipelineSchedules capability.
func AsPipelineScheduleClient(c Client, host string) (PipelineScheduleClient, error) {
	return requireFeature[PipelineScheduleClient](c, host, specFor(FeaturePipelineSchedules))
}
