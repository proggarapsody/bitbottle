package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireCloudRepo struct {
	FullName    string `json:"full_name"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	SCM         string `json:"scm"`
	Description string `json:"description"`
	Links       struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

func (w wireCloudRepo) toDomain() backend.Repository {
	ns := ""
	slug := w.Slug
	if parts := strings.SplitN(w.FullName, "/", 2); len(parts) == 2 {
		ns = parts[0]
		slug = parts[1]
	}
	return backend.Repository{
		Slug:        slug,
		Name:        w.Name,
		Namespace:   ns,
		SCM:         w.SCM,
		WebURL:      w.Links.HTML.Href,
		Description: w.Description,
	}
}

func (c *Client) ListRepos(limit int) ([]backend.Repository, error) {
	user, err := c.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/repositories/%s?pagelen=%d", user.Slug, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Repository, error) {
		var page cloudPagedResponse[wireCloudRepo]
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
	var w wireCloudRepo
	if err := c.getJSON(fmt.Sprintf("/repositories/%s/%s", ns, slug), &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}

type wireCloudCreateRepo struct {
	SCM       string `json:"scm"`
	IsPrivate bool   `json:"is_private"`
	Name      string `json:"name"`
}

func (c *Client) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	body := wireCloudCreateRepo{
		SCM:       in.SCM,
		IsPrivate: !in.Public,
		Name:      in.Name,
	}
	var w wireCloudRepo
	if err := c.postJSON(fmt.Sprintf("/repositories/%s/%s", ns, in.Name), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/repositories/%s/%s", ns, slug))
}
