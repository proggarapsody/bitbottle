package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// GetBranchModel returns the effective branching model for a repository.
// GET /repositories/{workspace}/{slug}/branching-model
func (c *Client) GetBranchModel(ws, slug string) (backend.BranchModel, error) {
	if ws == "" || slug == "" {
		return backend.BranchModel{}, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/branching-model",
		url.PathEscape(ws), url.PathEscape(slug))
	var out backend.BranchModel
	if err := c.getJSON(path, &out); err != nil {
		return backend.BranchModel{}, err
	}
	return out, nil
}

// GetBranchModelSettings returns the editable branching model settings.
// GET /repositories/{workspace}/{slug}/branching-model/settings
func (c *Client) GetBranchModelSettings(ws, slug string) (backend.BranchModelSettings, error) {
	if ws == "" || slug == "" {
		return backend.BranchModelSettings{}, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/branching-model/settings",
		url.PathEscape(ws), url.PathEscape(slug))
	var out backend.BranchModelSettings
	if err := c.getJSON(path, &out); err != nil {
		return backend.BranchModelSettings{}, err
	}
	return out, nil
}

// UpdateBranchModelSettings updates the branching model settings.
// PUT /repositories/{workspace}/{slug}/branching-model/settings
func (c *Client) UpdateBranchModelSettings(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
	if ws == "" || slug == "" {
		return backend.BranchModelSettings{}, fmt.Errorf("workspace and repo required")
	}
	path := fmt.Sprintf("/repositories/%s/%s/branching-model/settings",
		url.PathEscape(ws), url.PathEscape(slug))
	var out backend.BranchModelSettings
	if err := c.putJSON(path, in, &out); err != nil {
		return backend.BranchModelSettings{}, err
	}
	return out, nil
}
