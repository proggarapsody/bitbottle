package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wire types for JSON deserialization from Bitbucket Data Center responses.

type wireRepository struct {
	ID      int    `json:"id"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Project struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"project"`
	ScmID string `json:"scmId"`
	State string `json:"state"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func (w wireRepository) toDomain() backend.Repository {
	webURL := ""
	if len(w.Links.Self) > 0 {
		webURL = w.Links.Self[0].Href
	}
	return backend.Repository{
		Slug:      w.Slug,
		Name:      w.Name,
		Namespace: w.Project.Key,
		SCM:       w.ScmID,
		WebURL:    webURL,
	}
}

// ListRepos lists repositories accessible to the authenticated user.
func (c *Client) ListRepos(limit int) ([]backend.Repository, error) {
	var page PagedResponse[wireRepository]
	if err := c.getJSON(fmt.Sprintf("/repos?limit=%d", limit), &page); err != nil {
		return nil, err
	}
	repos := make([]backend.Repository, 0, len(page.Values))
	for _, w := range page.Values {
		repos = append(repos, w.toDomain())
	}
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	return repos, nil
}

// GetRepo fetches a single repository.
func (c *Client) GetRepo(ns, slug string) (backend.Repository, error) {
	var w wireRepository
	if err := c.getJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}

type wireCreateRepoInput struct {
	Name        string `json:"name"`
	ScmID       string `json:"scmId"`
	Public      bool   `json:"public"`
	Description string `json:"description,omitempty"`
}

// CreateRepo creates a new repository in ns.
func (c *Client) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	body := wireCreateRepoInput{
		Name:        in.Name,
		ScmID:       in.SCM,
		Public:      in.Public,
		Description: in.Description,
	}
	var w wireRepository
	if err := c.postJSON(fmt.Sprintf("/projects/%s/repos", ns), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}

// DeleteRepo deletes a repository.
func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), nil)
}

type wireAppProperties struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	DisplayName string `json:"displayName"`
}

// GetApplicationProperties fetches Bitbucket version information.
func (c *Client) GetApplicationProperties() (backend.AppProperties, error) {
	var w wireAppProperties
	if err := c.getJSON("/application-properties", &w); err != nil {
		return backend.AppProperties{}, err
	}
	return backend.AppProperties{
		Version:     w.Version,
		BuildNumber: w.BuildNumber,
		DisplayName: w.DisplayName,
	}, nil
}
