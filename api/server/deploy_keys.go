package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toDeployKeyDomain(w servergen.RestDeployKey) backend.DeployKey {
	return backend.DeployKey{
		ID:    w.ID,
		Label: w.Label,
		Key:   w.Key.Text,
	}
}

// ListDeployKeys returns all deploy keys for a repository.
// GET /rest/api/1.0/projects/{ns}/repos/{slug}/ssh
func (c *Client) ListDeployKeys(ns, slug string) ([]backend.DeployKey, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/ssh", ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DeployKey, error) {
		var page PagedResponse[servergen.RestDeployKey]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DeployKey, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toDeployKeyDomain(w))
		}
		return out, nil
	}, 0)
}

// AddDeployKey adds a deploy key to a repository.
// POST /rest/api/1.0/projects/{ns}/repos/{slug}/ssh
func (c *Client) AddDeployKey(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
	if ns == "" || slug == "" {
		return backend.DeployKey{}, fmt.Errorf("project and repo required")
	}
	if input.Key == "" {
		return backend.DeployKey{}, fmt.Errorf("key required")
	}
	body := servergen.RestAddDeployKeyBody{
		Key: servergen.RestSSHKeyInput{
			Text:  input.Key,
			Label: input.Label,
		},
	}
	var w servergen.RestDeployKey
	path := fmt.Sprintf("/projects/%s/repos/%s/ssh", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.DeployKey{}, err
	}
	return toDeployKeyDomain(w), nil
}

// DeleteDeployKey removes a deploy key from a repository.
// DELETE /rest/api/1.0/projects/{ns}/repos/{slug}/ssh/{id}
func (c *Client) DeleteDeployKey(ns, slug string, id int) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/ssh/%d", ns, slug, id)
	return c.delete(path, nil)
}
