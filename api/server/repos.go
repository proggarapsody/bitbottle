package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

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

func (c *Client) ListRepos(limit int) ([]backend.Repository, error) {
	path := fmt.Sprintf("/repos?limit=%d", limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Repository, error) {
		var page PagedResponse[wireRepository]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Repository, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

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

func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), nil)
}

type wireAppProperties struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	DisplayName string `json:"displayName"`
}

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
