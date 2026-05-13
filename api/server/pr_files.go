package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListPRFiles returns the files changed in a pull request, paginated.
// Server endpoint: GET /rest/api/1.0/projects/{ns}/repos/{slug}/pull-requests/{id}/changes
func (c *Client) ListPRFiles(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/changes", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DiffStatEntry, error) {
		var page PagedResponse[wireServerCommitChange]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DiffStatEntry, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}
