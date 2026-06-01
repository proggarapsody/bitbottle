package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// syncRepoRequest is the POST body for the merge-upstream endpoint.
type syncRepoRequest struct {
	Branch string `json:"branch,omitempty"`
}

// syncRepoResponse is the wire response from the merge-upstream endpoint.
type syncRepoResponse struct {
	Behind        int `json:"behind"`
	CommitsMerged int `json:"commits_merged"`
}

// SyncRepo synchronises a fork branch with its upstream.
// POST /repositories/{workspace}/{slug}/merge-upstream
// On 409 the transport returns ErrConflict (diverged history).
// On 404 the transport returns ErrNotFound.
func (c *Client) SyncRepo(ns, slug, branch string) (backend.SyncResult, error) {
	body := syncRepoRequest{Branch: branch}
	path := fmt.Sprintf("/repositories/%s/%s/merge-upstream",
		url.PathEscape(ns), url.PathEscape(slug))

	var resp syncRepoResponse
	if err := c.postJSON(path, body, &resp); err != nil {
		return backend.SyncResult{}, err
	}
	return backend.SyncResult{
		Behind:        resp.Behind,
		CommitsMerged: resp.CommitsMerged,
	}, nil
}
