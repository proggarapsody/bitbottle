package backend

import "fmt"

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
	sc, ok := c.(PipelineScheduleClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePipelineSchedules),
			Message: fmt.Sprintf("pipeline schedules are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return sc, nil
}
