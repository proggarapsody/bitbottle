package cloud

import (
	"fmt"

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

// ListCommitStatuses lists build / CI statuses reported against a commit hash.
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
