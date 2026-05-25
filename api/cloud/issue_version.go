package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type cloudVersion struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func toVersionDomain(v cloudVersion) backend.IssueVersion {
	return backend.IssueVersion{
		ID:   v.ID,
		Name: v.Name,
	}
}

// ListIssueVersions returns all issue versions for a repository.
func (c *Client) ListIssueVersions(ns, slug string, limit int) ([]backend.IssueVersion, error) {
	path := fmt.Sprintf("/repositories/%s/%s/versions",
		url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.IssueVersion, error) {
		var page cloudPagedResponse[cloudVersion]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.IssueVersion, 0, len(page.Values))
		for _, v := range page.Values {
			out = append(out, toVersionDomain(v))
		}
		return out, nil
	}, limit)
}

// GetIssueVersion returns a single issue version by ID.
func (c *Client) GetIssueVersion(ns, slug string, id int) (backend.IssueVersion, error) {
	var v cloudVersion
	path := fmt.Sprintf("/repositories/%s/%s/versions/%d",
		url.PathEscape(ns), url.PathEscape(slug), id)
	if err := c.getJSON(path, &v); err != nil {
		return backend.IssueVersion{}, err
	}
	return toVersionDomain(v), nil
}

// CreateIssueVersion creates a new issue version on a repository.
func (c *Client) CreateIssueVersion(ns, slug, name string) (backend.IssueVersion, error) {
	body := map[string]string{"name": name}
	var v cloudVersion
	path := fmt.Sprintf("/repositories/%s/%s/versions",
		url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &v); err != nil {
		return backend.IssueVersion{}, err
	}
	return toVersionDomain(v), nil
}

// DeleteIssueVersion deletes an issue version by ID.
func (c *Client) DeleteIssueVersion(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/versions/%d",
		url.PathEscape(ns), url.PathEscape(slug), id)
	return c.http.DeleteJSON(path, nil)
}
