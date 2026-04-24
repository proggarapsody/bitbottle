package cloud

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudRepo struct {
	FullName string `json:"full_name"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	SCM      string `json:"scm"`
	Links    struct {
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
		Slug:      slug,
		Name:      w.Name,
		Namespace: ns,
		SCM:       w.SCM,
		WebURL:    w.Links.HTML.Href,
	}
}

// ListRepos lists repositories visible to the authenticated user.
// Bitbucket Cloud's /repositories endpoint is not workspace-scoped here;
// callers that need workspace-scoped listing should use GetRepo with ns.
func (c *Client) ListRepos(limit int) ([]backend.Repository, error) {
	var page cloudPagedResponse[wireCloudRepo]
	path := fmt.Sprintf("/repositories?pagelen=%d", limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	repos := make([]backend.Repository, 0, len(page.Values))
	for _, w := range page.Values {
		repos = append(repos, w.toDomain())
	}
	return repos, nil
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
