package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toCommitStatusDomain(w cloudgen.CloudCommitStatus) backend.CommitStatus {
	return backend.CommitStatus{
		Key:         w.Key,
		State:       w.State,
		Name:        w.Name,
		Description: w.Description,
		URL:         w.URL,
	}
}

func (c *Client) ReportCommitStatus(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses/build",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(hash))
	body := cloudgen.CloudCommitStatusBody{Key: input.Key, State: input.State, URL: input.URL, Name: input.Name, Description: input.Description}
	var w cloudgen.CloudCommitStatus
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.CommitStatus{}, err
	}
	return toCommitStatusDomain(w), nil
}

func (c *Client) ListCommitStatuses(ns, slug, hash string) ([]backend.CommitStatus, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(hash))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.CommitStatus, error) {
		var page cloudPagedResponse[cloudgen.CloudCommitStatus]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CommitStatus, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitStatusDomain(w))
		}
		return out, nil
	}, 0)
}
