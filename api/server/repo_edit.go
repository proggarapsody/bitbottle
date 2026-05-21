package server

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// wireServerEditRepoInput is the Server wire shape for
// PUT /rest/api/1.0/projects/{ns}/repos/{slug}.
// Server only supports description; all other fields from EditRepoInput
// are silently ignored (not an error — they are simply not sent).
type wireServerEditRepoInput struct {
	Description *string `json:"description,omitempty"`
}

// EditRepo updates mutable repository metadata on a Bitbucket Server / Data
// Center instance. Only Description is forwarded; Cloud-only fields
// (Website, Language, ForkPolicy, HasIssues, HasWiki) are silently dropped.
func (c *Client) EditRepo(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
	body := wireServerEditRepoInput{
		Description: in.Description,
	}
	var w servergen.RestRepository
	path := fmt.Sprintf("/projects/%s/repos/%s", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepositoryDomain(w), nil
}
