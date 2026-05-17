package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const schedulesPath = "/repositories/%s/%s/pipelines_config/schedules"
const schedulePath = "/repositories/%s/%s/pipelines_config/schedules/%s"

func toPipelineScheduleDomain(w cloudgen.CloudPipelineSchedule) backend.PipelineSchedule {
	branch := w.Target.Branch
	if branch == "" {
		branch = w.Target.RefName
	}
	return backend.PipelineSchedule{
		UUID:           stripBraces(w.UUID),
		Enabled:        w.Enabled,
		CronExpression: w.CronExpression,
		Branch:         branch,
	}
}

// ListPipelineSchedules returns all pipeline schedules for a repository.
func (c *Client) ListPipelineSchedules(ns, slug string) ([]backend.PipelineSchedule, error) {
	path := fmt.Sprintf(schedulesPath, ns, slug)
	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.PipelineSchedule, error) {
			var page struct {
				Values []cloudgen.CloudPipelineSchedule `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.PipelineSchedule, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, toPipelineScheduleDomain(w))
			}
			return out, nil
		},
		0, // unbounded
	)
}

// CreatePipelineSchedule creates a new pipeline schedule.
func (c *Client) CreatePipelineSchedule(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
	body := cloudgen.CloudPipelineScheduleCreate{
		Enabled:        input.Enabled,
		CronExpression: input.CronExpression,
		Target: cloudgen.CloudScheduleTarget{
			RefType: "branch",
			RefName: input.Branch,
			Type:    "pipeline_ref_target",
		},
	}
	var w cloudgen.CloudPipelineSchedule
	path := fmt.Sprintf(schedulesPath, ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PipelineSchedule{}, err
	}
	return toPipelineScheduleDomain(w), nil
}

// DeletePipelineSchedule deletes a pipeline schedule by UUID.
func (c *Client) DeletePipelineSchedule(ns, slug, uuid string) error {
	path := fmt.Sprintf(schedulePath, ns, slug, braceUUID(uuid))
	return c.delete(path)
}
