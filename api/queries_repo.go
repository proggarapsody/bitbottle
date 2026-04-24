package api

import "fmt"

// Repository represents a Bitbucket repository.
type Repository struct {
	ID      int       `json:"id"`
	Slug    string    `json:"slug"`
	Name    string    `json:"name"`
	Project Project   `json:"project"`
	ScmID   string    `json:"scmId"`
	State   string    `json:"state"`
	Public  bool      `json:"public"`
	Links   RepoLinks `json:"links"`
}

// Project is a Bitbucket project key+name.
type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// RepoLinks holds clone and web links.
type RepoLinks struct {
	Clone []CloneLink `json:"clone"`
	Self  []SelfLink  `json:"self"`
}

// CloneLink is one clone URL entry.
type CloneLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

// SelfLink is a web UI link.
type SelfLink struct {
	Href string `json:"href"`
}

// AppProperties holds Bitbucket version info.
type AppProperties struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	DisplayName string `json:"displayName"`
}

// ListRepos lists repositories accessible to the authenticated user.
func (c *Client) ListRepos(limit int) ([]Repository, error) {
	var page PagedResponse[Repository]
	if err := c.GetJSON(fmt.Sprintf("/repos?limit=%d", limit), &page); err != nil {
		return nil, err
	}
	repos := page.Values
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	return repos, nil
}

// GetRepo fetches a single repository.
func (c *Client) GetRepo(project, slug string) (Repository, error) {
	var repo Repository
	if err := c.GetJSON(fmt.Sprintf("/projects/%s/repos/%s", project, slug), &repo); err != nil {
		return Repository{}, err
	}
	return repo, nil
}

// CreateRepoInput is the request body for repo creation.
type CreateRepoInput struct {
	Name        string `json:"name"`
	ScmID       string `json:"scmId"`
	Public      bool   `json:"public"`
	Description string `json:"description,omitempty"`
}

// CreateRepo creates a new repository in project.
func (c *Client) CreateRepo(project string, input CreateRepoInput) (Repository, error) {
	var repo Repository
	if err := c.PostJSON(fmt.Sprintf("/projects/%s/repos", project), input, &repo); err != nil {
		return Repository{}, err
	}
	return repo, nil
}

// DeleteRepo deletes a repository.
func (c *Client) DeleteRepo(project, slug string) error {
	return c.Delete(fmt.Sprintf("/projects/%s/repos/%s", project, slug))
}

// GetApplicationProperties fetches Bitbucket version info.
func (c *Client) GetApplicationProperties() (AppProperties, error) {
	var props AppProperties
	if err := c.GetJSON("/application-properties", &props); err != nil {
		return AppProperties{}, err
	}
	return props, nil
}
