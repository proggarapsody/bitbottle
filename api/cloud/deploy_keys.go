package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireCloudDeployKey is the Cloud wire shape for a deploy key.
type wireCloudDeployKey struct {
	ID       int    `json:"id"`
	Label    string `json:"label"`
	Key      string `json:"key"`
	ReadOnly bool   `json:"read_only"`
}

func (w wireCloudDeployKey) toDomain() backend.DeployKey {
	return backend.DeployKey{
		ID:       w.ID,
		Label:    w.Label,
		Key:      w.Key,
		ReadOnly: w.ReadOnly,
	}
}

// ListDeployKeys returns all deploy keys for a repository.
// GET /repositories/{workspace}/{slug}/deploy-keys
func (c *Client) ListDeployKeys(ns, slug string) ([]backend.DeployKey, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys", url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DeployKey, error) {
		var page cloudPagedResponse[wireCloudDeployKey]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DeployKey, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// addDeployKeyBody is the request body for creating a deploy key.
type addDeployKeyBody struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// AddDeployKey adds a deploy key to a repository.
// POST /repositories/{workspace}/{slug}/deploy-keys
func (c *Client) AddDeployKey(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
	if ns == "" || slug == "" {
		return backend.DeployKey{}, fmt.Errorf("workspace and repo required")
	}
	if input.Key == "" {
		return backend.DeployKey{}, fmt.Errorf("key required")
	}
	body := addDeployKeyBody{Key: input.Key, Label: input.Label}
	var w wireCloudDeployKey
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.DeployKey{}, err
	}
	return w.toDomain(), nil
}

// DeleteDeployKey removes a deploy key from a repository.
// DELETE /repositories/{workspace}/{slug}/deploy-keys/{id}
func (c *Client) DeleteDeployKey(ns, slug string, id int) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/%d", url.PathEscape(ns), url.PathEscape(slug), id)
	return c.delete(path)
}
