package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

// wireEditRepoInput is the Cloud wire shape for PUT /repositories/{ws}/{slug}.
// Only non-nil fields are serialised (omitempty).
type wireEditRepoInput struct {
	Description *string `json:"description,omitempty"`
	Website     *string `json:"website,omitempty"`
	Language    *string `json:"language,omitempty"`
	ForkPolicy  *string `json:"fork_policy,omitempty"`
	HasIssues   *bool   `json:"has_issues,omitempty"`
	HasWiki     *bool   `json:"has_wiki,omitempty"`
}

// EditRepo updates mutable metadata fields on a Bitbucket Cloud repository.
// Nil pointer fields in in are left unchanged.
func (c *Client) EditRepo(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
	body := wireEditRepoInput{
		Description: in.Description,
		Website:     in.Website,
		Language:    in.Language,
		ForkPolicy:  in.ForkPolicy,
		HasIssues:   in.HasIssues,
		HasWiki:     in.HasWiki,
	}
	var w cloudgen.CloudRepo
	path := fmt.Sprintf("/repositories/%s/%s", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepoDomain(w), nil
}
