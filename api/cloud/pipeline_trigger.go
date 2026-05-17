package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

// TriggerPipeline implements backend.PipelineTriggerClient for Bitbucket Cloud.
func (c *Client) TriggerPipeline(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
	body := cloudgen.CloudTriggerBody{
		Target: cloudgen.CloudTriggerTarget{
			RefType: "branch",
			Type:    "pipeline_ref_target",
			RefName: input.Branch,
		},
	}
	if len(input.Variables) > 0 {
		vars := make([]cloudgen.CloudTriggerVarItem, 0, len(input.Variables))
		for _, v := range input.Variables {
			vars = append(vars, cloudgen.CloudTriggerVarItem{
				Key:     v.Key,
				Value:   v.Value,
				Secured: v.Secured,
			})
		}
		body.Variables = &vars
	}

	var resp cloudgen.CloudTriggerResponse
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/", ns, slug)
	if err := c.postJSON(path, body, &resp); err != nil {
		return backend.PipelineTriggerResult{}, err
	}

	link := ""
	if len(resp.Links.Self) > 0 {
		link = resp.Links.Self[0].Href
	}
	return backend.PipelineTriggerResult{
		UUID:  stripBraces(resp.UUID),
		State: resp.State.Name,
		Link:  link,
	}, nil
}
