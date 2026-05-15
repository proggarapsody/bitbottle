package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const cachesPath = "/repositories/%s/%s/pipelines_config/caches/"
const cachePath = "/repositories/%s/%s/pipelines_config/caches/%s"

type wireCloudPipelineCache struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	CreatedOn     string `json:"created_on"`
}

func (w wireCloudPipelineCache) toDomain() backend.PipelineCache {
	return backend.PipelineCache{
		UUID:          stripBraces(w.UUID),
		Name:          w.Name,
		Path:          w.Path,
		FileSizeBytes: w.FileSizeBytes,
		CreatedOn:     w.CreatedOn,
	}
}

// ListPipelineCaches returns all pipeline caches for a repository.
func (c *Client) ListPipelineCaches(ns, slug string) ([]backend.PipelineCache, error) {
	path := fmt.Sprintf(cachesPath, ns, slug)
	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.PipelineCache, error) {
			var page struct {
				Values []wireCloudPipelineCache `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.PipelineCache, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, w.toDomain())
			}
			return out, nil
		},
		0, // unbounded
	)
}

// DeletePipelineCache deletes a pipeline cache by UUID.
func (c *Client) DeletePipelineCache(ns, slug, uuid string) error {
	path := fmt.Sprintf(cachePath, ns, slug, braceUUID(uuid))
	return c.delete(path)
}
