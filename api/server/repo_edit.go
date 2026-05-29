package server

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// EditRepo updates mutable repository metadata on a Bitbucket Server / Data
// Center instance. Only Description is forwarded; Cloud-only fields
// (Website, Language, ForkPolicy, HasIssues, HasWiki) are silently dropped.
func (c *Client) EditRepo(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
	body := servergen.RestEditRepoInput{
		Description: in.Description,
	}
	var w servergen.RestRepository
	path := fmt.Sprintf("/projects/%s/repos/%s", url.PathEscape(ns), url.PathEscape(slug))
	// OCC: not required — Bitbucket Server repo metadata PUT does not version entity
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepositoryDomain(w), nil
}
