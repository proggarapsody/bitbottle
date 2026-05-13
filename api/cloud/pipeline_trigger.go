package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireTriggerBody is the JSON body sent to POST /repositories/{ws}/{slug}/pipelines/.
type wireTriggerBody struct {
	Target    wireTriggerTarget    `json:"target"`
	Variables []wireTriggerVarItem `json:"variables,omitempty"`
}

type wireTriggerTarget struct {
	RefType string `json:"ref_type"`
	Type    string `json:"type"`
	RefName string `json:"ref_name"`
}

type wireTriggerVarItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Secured bool   `json:"secured"`
}

// wireTriggerResponse is the relevant subset of the Cloud pipeline response.
type wireTriggerResponse struct {
	UUID  string `json:"uuid"`
	State struct {
		Name string `json:"name"`
	} `json:"state"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

// TriggerPipeline implements backend.PipelineTriggerClient for Bitbucket Cloud.
func (c *Client) TriggerPipeline(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
	body := wireTriggerBody{
		Target: wireTriggerTarget{
			RefType: "branch",
			Type:    "pipeline_ref_target",
			RefName: input.Branch,
		},
	}
	for _, v := range input.Variables {
		body.Variables = append(body.Variables, wireTriggerVarItem{
			Key:     v.Key,
			Value:   v.Value,
			Secured: v.Secured,
		})
	}

	var resp wireTriggerResponse
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
