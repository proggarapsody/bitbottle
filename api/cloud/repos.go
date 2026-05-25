package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toRepoDomain(w cloudgen.CloudRepo) backend.Repository {
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
		SCM:         w.Scm,
		WebURL:      w.Links.HTML.Href,
		Description: w.Description,
		IsPrivate:   w.IsPrivate,
	}
}

func (c *Client) ListRepos(ns string, limit int) ([]backend.Repository, error) {
	if ns == "" {
		return nil, fmt.Errorf("workspace required for Bitbucket Cloud; use: repo list WORKSPACE")
	}
	path := fmt.Sprintf("/repositories/%s?pagelen=%d", ns, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Repository, error) {
		var page cloudPagedResponse[cloudgen.CloudRepo]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Repository, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRepoDomain(w))
		}
		return out, nil
	}, limit)
}

func (c *Client) GetRepo(ns, slug string) (backend.Repository, error) {
	type cloneLink struct {
		Name string `json:"name"`
		Href string `json:"href"`
	}
	var w struct {
		Description string `json:"description"`
		FullName    string `json:"full_name"`
		IsPrivate   bool   `json:"is_private"`
		Name        string `json:"name"`
		Scm         string `json:"scm"`
		Slug        string `json:"slug"`
		Links       struct {
			HTML  cloudgen.CloudHTMLLink `json:"html"`
			Clone []cloneLink            `json:"clone"`
		} `json:"links"`
	}
	if err := c.getJSON(fmt.Sprintf("/repositories/%s/%s", ns, slug), &w); err != nil {
		return backend.Repository{}, stampRepoNotFound(err, ns, slug)
	}
	ns2, slug2 := ns, w.Slug
	if parts := strings.SplitN(w.FullName, "/", 2); len(parts) == 2 {
		ns2, slug2 = parts[0], parts[1]
	}
	repo := backend.Repository{
		Slug: slug2, Name: w.Name, Namespace: ns2, SCM: w.Scm,
		WebURL: w.Links.HTML.Href, Description: w.Description, IsPrivate: w.IsPrivate,
	}
	for _, cl := range w.Links.Clone {
		repo.CloneURLs = append(repo.CloneURLs, backend.CloneURL{Name: cl.Name, URL: cl.Href})
	}
	return repo, nil
}

// stampRepoNotFound annotates a 404-on-repo error with CodeRepoNotFound +
// Resource/ID. Other statuses pass through so adapter errors keep their
// generic classification (e.g. 401 → CodeAuthInvalidToken).
func stampRepoNotFound(err error, ns, slug string) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 404 {
		return err
	}
	return backend.StampCode(err, backend.CodeRepoNotFound, "repository", ns+"/"+slug, "")
}

func (c *Client) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	body := cloudgen.CloudCreateRepo{
		Scm:       in.SCM,
		IsPrivate: !in.Public,
		Name:      in.Name,
	}
	var w cloudgen.CloudRepo
	if err := c.postJSON(fmt.Sprintf("/repositories/%s/%s", ns, in.Name), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepoDomain(w), nil
}

func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/repositories/%s/%s", ns, slug))
}

func (c *Client) RenameRepo(ns, slug, newName string) (backend.Repository, error) {
	body := map[string]string{"name": newName}
	var w cloudgen.CloudRepo
	if err := c.putJSON(fmt.Sprintf("/repositories/%s/%s", ns, slug), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepoDomain(w), nil
}

func (c *Client) SetRepoVisibility(ns, slug string, isPrivate bool) error {
	body := struct {
		IsPrivate bool `json:"is_private"`
	}{IsPrivate: isPrivate}
	var ignore cloudgen.CloudRepo
	return c.putJSON(fmt.Sprintf("/repositories/%s/%s", ns, slug), body, &ignore)
}

func (c *Client) ForkRepo(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
	body := cloudgen.CloudForkBody{
		Workspace: cloudgen.CloudForkWorkspace{Slug: in.Workspace},
	}
	if in.Name != "" {
		body.Name = &in.Name
	}
	var w cloudgen.CloudRepo
	if err := c.postJSON(fmt.Sprintf("/repositories/%s/%s/forks", ns, slug), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepoDomain(w), nil
}
