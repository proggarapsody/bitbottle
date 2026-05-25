package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

// cloudPipelineRaw is a minimal struct for reading the source pipeline's
// target fields including the commit hash, which is not exposed in the
// generated CloudPipelineTarget.
type cloudPipelineRaw struct {
	Target struct {
		Type    string `json:"type"`
		RefName string `json:"ref_name"`
		RefType string `json:"ref_type"`
		Commit  struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"target"`
}

// cloudRerunBody is the POST body for triggering a pipeline rerun.
type cloudRerunBody struct {
	Target cloudRerunTarget `json:"target"`
}

type cloudRerunTarget struct {
	Type    string            `json:"type"`
	RefName string            `json:"ref_name"`
	RefType string            `json:"ref_type,omitempty"`
	Commit  *cloudRerunCommit `json:"commit,omitempty"`
}

type cloudRerunCommit struct {
	Type string `json:"type"`
	Hash string `json:"hash"`
}

// RerunPipeline fetches the source pipeline's target (including commit hash)
// and triggers a new run at exactly the same commit. When the source pipeline
// has no commit hash (e.g. custom pipelines), the trigger falls back to a
// ref-only request.
func (c *Client) RerunPipeline(ns, slug, sourceUUID string) (backend.Pipeline, error) {
	var raw cloudPipelineRaw
	getPath := fmt.Sprintf("/repositories/%s/%s/pipelines/%s", ns, slug, braceUUID(sourceUUID))
	if err := c.getJSON(getPath, &raw); err != nil {
		return backend.Pipeline{}, err
	}

	targetType := raw.Target.Type
	if targetType == "" {
		targetType = "pipeline_ref_target"
	}

	body := cloudRerunBody{
		Target: cloudRerunTarget{
			Type:    targetType,
			RefName: raw.Target.RefName,
			RefType: raw.Target.RefType,
		},
	}
	if raw.Target.Commit.Hash != "" {
		body.Target.Commit = &cloudRerunCommit{Type: "commit", Hash: raw.Target.Commit.Hash}
	}

	var resp cloudgen.CloudPipeline
	postPath := fmt.Sprintf("/repositories/%s/%s/pipelines/", ns, slug)
	if err := c.postJSON(postPath, body, &resp); err != nil {
		return backend.Pipeline{}, err
	}
	return toPipelineDomain(resp, ns, slug), nil
}
