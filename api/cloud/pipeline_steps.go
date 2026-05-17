package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListPipelineSteps lists the steps in a pipeline run.
func (c *Client) ListPipelineSteps(ns, slug, uuid string) ([]backend.PipelineStep, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/", ns, slug, braceUUID(uuid))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineStep, error) {
		var page cloudPagedResponse[cloudgen.CloudPipelineStep]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		steps := make([]backend.PipelineStep, 0, len(page.Values))
		for _, w := range page.Values {
			steps = append(steps, toPipelineStepDomain(w))
		}
		return steps, nil
	}, 0)
}

func toPipelineStepDomain(w cloudgen.CloudPipelineStep) backend.PipelineStep {
	state := w.State.Name
	result := w.State.Result.Name
	if state == "COMPLETED" && result != "" {
		state = result
	}
	return backend.PipelineStep{
		UUID:     stripBraces(w.UUID),
		Name:     w.Name,
		State:    state,
		Result:   result,
		Duration: w.DurationInSeconds,
	}
}
