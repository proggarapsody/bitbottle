package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toRepositoryDomain(w servergen.RestRepository) backend.Repository {
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
		ID:        w.ID,
		IsPrivate: !w.Public,
	}
}

// ListRepos lists all repositories accessible to the authenticated user.
// ns is ignored for Bitbucket Server (the REST API lists across all projects).
func (c *Client) ListRepos(_ string, limit int) ([]backend.Repository, error) {
	path := fmt.Sprintf("/repos?limit=%d", limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Repository, error) {
		var page PagedResponse[servergen.RestRepository]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Repository, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRepositoryDomain(w))
		}
		return out, nil
	}, limit)
}

func (c *Client) GetRepo(ns, slug string) (backend.Repository, error) {
	var w servergen.RestRepository
	if err := c.getJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), &w); err != nil {
		return backend.Repository{}, stampRepoNotFound(err, ns, slug)
	}
	return toRepositoryDomain(w), nil
}

// stampRepoNotFound annotates a 404-on-repo error with CodeRepoNotFound +
// Resource/ID. Server uses PROJECT/REPO casing where PROJECT is the project
// key, not name — the catalogue hint mentions this distinction.
func stampRepoNotFound(err error, ns, slug string) error {
	var de *backend.DomainError
	if !errors.As(err, &de) || de.HTTPStatus() != 404 {
		return err
	}
	return backend.StampCode(err, backend.CodeRepoNotFound, "repository", ns+"/"+slug, "")
}

func (c *Client) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	body := servergen.RestCreateRepoInput{
		Name:        in.Name,
		ScmID:       in.SCM,
		Public:      in.Public,
		Description: in.Description,
	}
	var w servergen.RestRepository
	if err := c.postJSON(fmt.Sprintf("/projects/%s/repos", ns), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepositoryDomain(w), nil
}

func (c *Client) DeleteRepo(ns, slug string) error {
	return c.delete(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), nil)
}

func (c *Client) RenameRepo(ns, slug, newName string) (backend.Repository, error) {
	body := map[string]string{"name": newName}
	var w servergen.RestRepository
	if err := c.putJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepositoryDomain(w), nil
}

func (c *Client) SetRepoVisibility(ns, slug string, isPrivate bool) error {
	body := struct {
		Public bool `json:"public"`
	}{Public: !isPrivate}
	var ignore servergen.RestRepository
	return c.putJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), body, &ignore)
}
