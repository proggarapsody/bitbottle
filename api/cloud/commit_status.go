package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudCommitStatus struct {
	Key         string `json:"key"`
	State       string `json:"state"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func (w wireCloudCommitStatus) toDomain() backend.CommitStatus {
	return backend.CommitStatus{
		Key:         w.Key,
		State:       w.State,
		Name:        w.Name,
		Description: w.Description,
		URL:         w.URL,
	}
}

type wireCommitStatusBody struct {
	Key         string `json:"key"`
	State       string `json:"state"`
	URL         string `json:"url,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) ReportCommitStatus(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses/build",
		url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(hash))
	body := wireCommitStatusBody{Key: input.Key, State: input.State, URL: input.URL, Name: input.Name, Description: input.Description}
	var w wireCloudCommitStatus
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.CommitStatus{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) ListCommitStatuses(ns, slug, hash string) ([]backend.CommitStatus, error) {
	var page cloudPagedResponse[wireCloudCommitStatus]
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses?pagelen=100", ns, slug, hash)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.CommitStatus, 0, len(page.Values))
	for _, w := range page.Values {
		out = append(out, w.toDomain())
	}
	return out, nil
}
