package cloud

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudPipeline struct {
	UUID        string `json:"uuid"`
	BuildNumber int    `json:"build_number"`
	State       struct {
		Name   string `json:"name"`
		Result struct {
			Name string `json:"name"`
		} `json:"result"`
	} `json:"state"`
	Target struct {
		RefType string `json:"ref_type"`
		RefName string `json:"ref_name"`
	} `json:"target"`
	CreatedOn         string `json:"created_on"`
	DurationInSeconds int    `json:"duration_in_seconds"`
	Links             struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func (w wireCloudPipeline) toDomain(ns, slug string) backend.Pipeline {
	state := w.State.Name
	if state == "COMPLETED" && w.State.Result.Name != "" {
		state = w.State.Result.Name
	}
	// Bitbucket Cloud's links.self points to the REST API, not the browser UI.
	// Construct the browser URL from workspace, repo slug, and build number.
	webURL := fmt.Sprintf(
		"https://bitbucket.org/%s/%s/pipelines/results/%d",
		ns, slug, w.BuildNumber,
	)
	return backend.Pipeline{
		UUID:        stripBraces(w.UUID),
		BuildNumber: w.BuildNumber,
		State:       state,
		RefType:     w.Target.RefType,
		RefName:     w.Target.RefName,
		CreatedOn:   w.CreatedOn,
		Duration:    w.DurationInSeconds,
		WebURL:      webURL,
	}
}

// ListPipelines lists recent pipeline runs for a repository.
func (c *Client) ListPipelines(ns, slug string, limit int) ([]backend.Pipeline, error) {
	var page cloudPagedResponse[wireCloudPipeline]
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/?sort=-created_on&pagelen=%d", ns, slug, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	pipelines := make([]backend.Pipeline, 0, len(page.Values))
	for _, w := range page.Values {
		pipelines = append(pipelines, w.toDomain(ns, slug))
	}
	return pipelines, nil
}

// GetPipeline fetches a single pipeline run by UUID.
// Bitbucket Cloud requires pipeline UUIDs to be enclosed in curly braces in
// the URL path (e.g. "{abc-123}"), so we normalise the caller-supplied uuid.
func (c *Client) GetPipeline(ns, slug, uuid string) (backend.Pipeline, error) {
	var w wireCloudPipeline
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s", ns, slug, braceUUID(uuid))
	if err := c.getJSON(path, &w); err != nil {
		return backend.Pipeline{}, err
	}
	return w.toDomain(ns, slug), nil
}

// braceUUID wraps a UUID in curly braces if it is not already wrapped, as
// required by the Bitbucket Cloud pipeline API.
func braceUUID(uuid string) string {
	if len(uuid) > 0 && uuid[0] == '{' {
		return uuid
	}
	return "{" + uuid + "}"
}

// stripBraces removes leading '{' and trailing '}' from a UUID string as
// returned by the Bitbucket Cloud API, so the domain model stores a plain UUID.
func stripBraces(uuid string) string {
	return strings.Trim(uuid, "{}")
}

type wireRunPipelineInput struct {
	Target wireRunPipelineTarget `json:"target"`
}

type wireRunPipelineTarget struct {
	Type    string `json:"type"`
	RefType string `json:"ref_type"`
	RefName string `json:"ref_name"`
}

// RunPipeline triggers a new pipeline run on a branch.
func (c *Client) RunPipeline(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
	body := wireRunPipelineInput{
		Target: wireRunPipelineTarget{
			Type:    "pipeline_ref_target",
			RefType: "branch",
			RefName: in.Branch,
		},
	}
	var w wireCloudPipeline
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Pipeline{}, err
	}
	return w.toDomain(ns, slug), nil
}
