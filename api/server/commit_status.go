package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireServerCommitStatus struct {
	Key         string `json:"key"`
	State       string `json:"state"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func (w wireServerCommitStatus) toDomain() backend.CommitStatus {
	return backend.CommitStatus{
		Key:         w.Key,
		State:       w.State,
		Name:        w.Name,
		Description: w.Description,
		URL:         w.URL,
	}
}

// ListCommitStatuses lists build / CI statuses reported against a commit hash.
// Bitbucket Server / Data Center exposes these on a separate REST root,
// /rest/build-status/1.0, rather than the regular /rest/api/1.0 base.
//
// The ns/slug arguments are unused by the Server build-status API (statuses
// are keyed only by commit hash) but are kept for interface symmetry.
func (c *Client) ListCommitStatuses(_, _, hash string) ([]backend.CommitStatus, error) {
	var page PagedResponse[wireServerCommitStatus]
	path := fmt.Sprintf("/commits/%s?limit=100", hash)
	if err := c.buildStatusHTTP.GetJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.CommitStatus, 0, len(page.Values))
	for _, w := range page.Values {
		out = append(out, w.toDomain())
	}
	return out, nil
}
