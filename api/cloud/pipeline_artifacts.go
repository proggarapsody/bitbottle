package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudArtifact is the wire shape of one artifact in the Cloud API response.
type cloudArtifact struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	Links     struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func toArtifactDomain(w cloudArtifact) backend.PipelineArtifact {
	return backend.PipelineArtifact{
		Name:      w.Name,
		SizeBytes: w.SizeBytes,
		URL:       w.Links.Self.Href,
	}
}

// ListPipelineArtifacts returns artifacts produced by a pipeline step.
func (c *Client) ListPipelineArtifacts(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error) {
	path := fmt.Sprintf(
		"/repositories/%s/%s/pipelines/%s/steps/%s/artifacts",
		url.PathEscape(ws), url.PathEscape(slug),
		url.PathEscape(braceUUID(pipelineUUID)),
		url.PathEscape(braceUUID(stepUUID)),
	)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineArtifact, error) {
		var page struct {
			Values []cloudArtifact `json:"values"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PipelineArtifact, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toArtifactDomain(w))
		}
		return out, nil
	}, limit)
}

// DownloadPipelineArtifact streams a named artifact from a pipeline step to out.
func (c *Client) DownloadPipelineArtifact(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
	path := fmt.Sprintf(
		"/repositories/%s/%s/pipelines/%s/steps/%s/artifacts/%s",
		url.PathEscape(ws), url.PathEscape(slug),
		url.PathEscape(braceUUID(pipelineUUID)),
		url.PathEscape(braceUUID(stepUUID)),
		url.PathEscape(name),
	)
	rc, err := c.http.GetStream(path)
	if err != nil {
		return err
	}
	defer rc.Close() //nolint:errcheck
	_, err = io.Copy(out, rc)
	return err
}
