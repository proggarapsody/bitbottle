package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// restProject is the wire representation of a Bitbucket Server/DC project.
type restProject struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Links       struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func toServerProjectDomain(w restProject) backend.ServerProject {
	webURL := ""
	if len(w.Links.Self) > 0 {
		webURL = w.Links.Self[0].Href
	}
	return backend.ServerProject{
		Key:         w.Key,
		Name:        w.Name,
		Description: w.Description,
		Public:      w.Public,
		WebURL:      webURL,
	}
}

// ListServerProjects returns projects on Bitbucket Server/DC, optionally
// filtered by a name prefix.
func (c *Client) ListServerProjects(filter string, limit int) ([]backend.ServerProject, error) {
	path := fmt.Sprintf("/projects?limit=%d", limit)
	if filter != "" {
		path += "&name=" + url.QueryEscape(filter)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.ServerProject, error) {
		var page PagedResponse[restProject]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.ServerProject, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toServerProjectDomain(w))
		}
		return out, nil
	}, limit)
}

// GetServerProject fetches a single project by its key.
func (c *Client) GetServerProject(key string) (backend.ServerProject, error) {
	var w restProject
	if err := c.getJSON(fmt.Sprintf("/projects/%s", url.PathEscape(key)), &w); err != nil {
		return backend.ServerProject{}, err
	}
	return toServerProjectDomain(w), nil
}

// CreateServerProject creates a new project on Bitbucket Server/DC.
func (c *Client) CreateServerProject(in backend.CreateServerProjectInput) (backend.ServerProject, error) {
	type createRequest struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Type        string `json:"type"`
		Public      bool   `json:"public"`
	}
	body := createRequest{
		Key:         in.Key,
		Name:        in.Name,
		Description: in.Description,
		Type:        "NORMAL",
		Public:      in.Public,
	}
	var w restProject
	if err := c.postJSON("/projects", body, &w); err != nil {
		return backend.ServerProject{}, err
	}
	return toServerProjectDomain(w), nil
}

// UpdateServerProject patches a project on Bitbucket Server/DC.
// Only non-nil fields in in are changed.
func (c *Client) UpdateServerProject(key string, in backend.UpdateServerProjectInput) (backend.ServerProject, error) {
	// Fetch current state first so we can fill in unchanged fields.
	current, err := c.GetServerProject(key)
	if err != nil {
		return backend.ServerProject{}, err
	}
	type updateRequest struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Public      bool   `json:"public"`
	}
	req := updateRequest{
		Key:         current.Key,
		Name:        current.Name,
		Description: current.Description,
		Type:        "NORMAL",
		Public:      current.Public,
	}
	if in.Name != nil {
		req.Name = *in.Name
	}
	if in.Description != nil {
		req.Description = *in.Description
	}
	if in.Public != nil {
		req.Public = *in.Public
	}
	var w restProject
	if err := c.putJSON(fmt.Sprintf("/projects/%s", url.PathEscape(key)), req, &w); err != nil {
		return backend.ServerProject{}, err
	}
	return toServerProjectDomain(w), nil
}

// DeleteServerProject deletes a project by key.
func (c *Client) DeleteServerProject(key string) error {
	return c.delete(fmt.Sprintf("/projects/%s", url.PathEscape(key)), nil)
}
