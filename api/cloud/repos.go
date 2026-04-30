package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
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

// ListRepos lists repositories in the authenticated user's workspace, following
// all pagination pages.
func (c *Client) ListRepos(limit int) ([]backend.Repository, error) {
	user, err := c.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	var repos []backend.Repository
	path := fmt.Sprintf("/repositories/%s?pagelen=%d", user.Slug, limit)
	err = c.http.GetAllJSON(path, func(body []byte) error {
		var page cloudPagedResponse[wireCloudRepo]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, w := range page.Values {
			repos = append(repos, w.toDomain())
		}
		return nil
	})
	return repos, err
}

// GetRepo fetches a single repository.
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

// CreateRepo creates a new repository in the workspace ns.
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

// DeleteRepo deletes a repository.
func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/repositories/%s/%s", ns, slug))
}
