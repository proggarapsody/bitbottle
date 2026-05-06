package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ListPipelineSteps lists the steps in a pipeline run.
func (c *Client) ListPipelineSteps(ns, slug, uuid string) ([]backend.PipelineStep, error) {
	var page cloudPagedResponse[wireCloudPipelineStep]
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/", ns, slug, braceUUID(uuid))
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	steps := make([]backend.PipelineStep, 0, len(page.Values))
	for _, w := range page.Values {
		steps = append(steps, w.toDomain())
	}
	return steps, nil
}

type wireCloudPipelineStep struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	State struct {
		Name   string `json:"name"`
		Result struct {
			Name string `json:"name"`
		} `json:"result"`
	} `json:"state"`
	DurationInSeconds int `json:"duration_in_seconds"`
}

func (w wireCloudPipelineStep) toDomain() backend.PipelineStep {
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
