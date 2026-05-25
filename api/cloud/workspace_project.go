package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type cloudProject struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

func toWorkspaceProjectDomain(p cloudProject) backend.WorkspaceProject {
	return backend.WorkspaceProject{
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		IsPrivate:   p.IsPrivate,
	}
}

// CreateWorkspaceProject creates a new project in a workspace.
func (c *Client) CreateWorkspaceProject(ws string, input backend.CreateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
	body := cloudProject{
		Key:         input.Key,
		Name:        input.Name,
		Description: input.Description,
		IsPrivate:   input.IsPrivate,
	}
	var p cloudProject
	path := fmt.Sprintf("/workspaces/%s/projects", url.PathEscape(ws))
	if err := c.postJSON(path, body, &p); err != nil {
		return backend.WorkspaceProject{}, err
	}
	return toWorkspaceProjectDomain(p), nil
}

// GetWorkspaceProject returns a single workspace project by key.
func (c *Client) GetWorkspaceProject(ws, key string) (backend.WorkspaceProject, error) {
	var p cloudProject
	path := fmt.Sprintf("/workspaces/%s/projects/%s",
		url.PathEscape(ws), url.PathEscape(key))
	if err := c.getJSON(path, &p); err != nil {
		return backend.WorkspaceProject{}, err
	}
	return toWorkspaceProjectDomain(p), nil
}

// UpdateWorkspaceProject applies non-nil fields from input to the existing project
// and PUTs the merged result.
func (c *Client) UpdateWorkspaceProject(ws, key string, input backend.UpdateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
	// Fetch current state.
	existing, err := c.GetWorkspaceProject(ws, key)
	if err != nil {
		return backend.WorkspaceProject{}, err
	}
	// Merge non-nil fields.
	body := cloudProject{
		Key:         existing.Key,
		Name:        existing.Name,
		Description: existing.Description,
		IsPrivate:   existing.IsPrivate,
	}
	if input.Name != nil {
		body.Name = *input.Name
	}
	if input.Description != nil {
		body.Description = *input.Description
	}
	if input.IsPrivate != nil {
		body.IsPrivate = *input.IsPrivate
	}
	var p cloudProject
	path := fmt.Sprintf("/workspaces/%s/projects/%s",
		url.PathEscape(ws), url.PathEscape(key))
	if err := c.putJSON(path, body, &p); err != nil {
		return backend.WorkspaceProject{}, err
	}
	return toWorkspaceProjectDomain(p), nil
}

// DeleteWorkspaceProject deletes a workspace project by key.
func (c *Client) DeleteWorkspaceProject(ws, key string) error {
	path := fmt.Sprintf("/workspaces/%s/projects/%s",
		url.PathEscape(ws), url.PathEscape(key))
	return c.http.DeleteJSON(path, nil)
}
