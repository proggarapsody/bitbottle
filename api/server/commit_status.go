package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toCommitStatusDomain(w servergen.RestCommitStatus) backend.CommitStatus {
	return backend.CommitStatus{
		Key:         w.Key,
		State:       w.State,
		Name:        w.Name,
		Description: w.Description,
		URL:         w.URL,
	}
}

// ReportCommitStatus posts a build status against a commit hash.
// Bitbucket Server / Data Center exposes this on the separate REST root,
// /rest/build-status/1.0. The ns/slug arguments are unused (statuses are
// keyed only by commit hash) but are kept for interface symmetry.
func (c *Client) ReportCommitStatus(_, _, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
	body := struct {
		Key         string `json:"key"`
		State       string `json:"state"`
		URL         string `json:"url,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}{Key: input.Key, State: input.State, URL: input.URL, Name: input.Name, Description: input.Description}
	path := fmt.Sprintf("/commits/%s", hash)
	if err := c.buildStatusHTTP.PostJSON(path, body, nil); err != nil {
		return backend.CommitStatus{}, err
	}
	return backend.CommitStatus{Key: input.Key, State: input.State, Name: input.Name, Description: input.Description, URL: input.URL}, nil
}

// ListCommitStatuses lists build / CI statuses reported against a commit hash.
// Bitbucket Server / Data Center exposes these on a separate REST root,
// /rest/build-status/1.0, rather than the regular /rest/api/1.0 base.
//
// The ns/slug arguments are unused by the Server build-status API (statuses
// are keyed only by commit hash) but are kept for interface symmetry.
func (c *Client) ListCommitStatuses(_, _, hash string) ([]backend.CommitStatus, error) {
	var page PagedResponse[servergen.RestCommitStatus]
	path := fmt.Sprintf("/commits/%s?limit=100", hash)
	if err := c.buildStatusHTTP.GetJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.CommitStatus, 0, len(page.Values))
	for _, w := range page.Values {
		out = append(out, toCommitStatusDomain(w))
	}
	return out, nil
}
