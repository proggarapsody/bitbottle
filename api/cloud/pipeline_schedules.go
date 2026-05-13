package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const schedulesPath = "/repositories/%s/%s/pipelines_config/schedules"
const schedulePath = "/repositories/%s/%s/pipelines_config/schedules/%s"

type wireCloudScheduleTarget struct {
	Branch string `json:"branch,omitempty"`
	// RefType and RefName are used when creating a schedule.
	RefType string `json:"ref_type,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Type    string `json:"type,omitempty"`
}

type wireCloudSchedule struct {
	UUID           string                  `json:"uuid"`
	Enabled        bool                    `json:"enabled"`
	CronExpression string                  `json:"cron_expression"`
	Target         wireCloudScheduleTarget `json:"target"`
}

func (w wireCloudSchedule) toDomain() backend.PipelineSchedule {
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

type wireCloudScheduleCreate struct {
	Enabled        bool                    `json:"enabled"`
	CronExpression string                  `json:"cron_expression"`
	Target         wireCloudScheduleTarget `json:"target"`
}

// ListPipelineSchedules returns all pipeline schedules for a repository.
func (c *Client) ListPipelineSchedules(ns, slug string) ([]backend.PipelineSchedule, error) {
	path := fmt.Sprintf(schedulesPath, ns, slug)
	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.PipelineSchedule, error) {
			var page struct {
				Values []wireCloudSchedule `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.PipelineSchedule, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, w.toDomain())
			}
			return out, nil
		},
		0, // unbounded
	)
}

// CreatePipelineSchedule creates a new pipeline schedule.
func (c *Client) CreatePipelineSchedule(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
	body := wireCloudScheduleCreate{
		Enabled:        input.Enabled,
		CronExpression: input.CronExpression,
		Target: wireCloudScheduleTarget{
			RefType: "branch",
			RefName: input.Branch,
			Type:    "pipeline_ref_target",
		},
	}
	var w wireCloudSchedule
	path := fmt.Sprintf(schedulesPath, ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PipelineSchedule{}, err
	}
	return w.toDomain(), nil
}

// DeletePipelineSchedule deletes a pipeline schedule by UUID.
func (c *Client) DeletePipelineSchedule(ns, slug, uuid string) error {
	path := fmt.Sprintf(schedulePath, ns, slug, braceUUID(uuid))
	return c.delete(path)
}
