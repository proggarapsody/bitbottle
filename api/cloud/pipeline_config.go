package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

const pipelinesConfigPath = "/repositories/%s/%s/pipelines_config"

// cloudPipelinesConfig is the wire representation of the Cloud pipelines_config response.
type cloudPipelinesConfig struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
}

// GetPipelinesConfig returns the pipeline configuration for a repository.
// GET /2.0/repositories/{workspace}/{repo_slug}/pipelines_config
func (c *Client) GetPipelinesConfig(ws, slug string) (backend.PipelineConfig, error) {
	path := fmt.Sprintf(pipelinesConfigPath, url.PathEscape(ws), url.PathEscape(slug))
	var w cloudPipelinesConfig
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineConfig{}, err
	}
	return backend.PipelineConfig{Enabled: w.Enabled}, nil
}

// UpdatePipelinesConfig updates the pipeline configuration for a repository.
// PUT /2.0/repositories/{workspace}/{repo_slug}/pipelines_config
func (c *Client) UpdatePipelinesConfig(ws, slug string, input backend.PipelineConfig) (backend.PipelineConfig, error) {
	path := fmt.Sprintf(pipelinesConfigPath, url.PathEscape(ws), url.PathEscape(slug))
	body := cloudPipelinesConfig{Enabled: input.Enabled}
	var w cloudPipelinesConfig
	if err := c.http.PutJSON(path, body, &w); err != nil {
		return backend.PipelineConfig{}, err
	}
	return backend.PipelineConfig{Enabled: w.Enabled}, nil
}
