package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListPRCommits returns the commits in a pull request, paginated.
func (c *Client) ListPRCommits(ns, slug string, prID int) ([]backend.Commit, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/commits?limit=100", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Commit, error) {
		var page PagedResponse[wireServerCommit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Commit, 0, len(page.Values))
		for _, w := range page.Values {
			commit := w.toDomain()
			commit.WebURL = c.commitWebURL(ns, slug, commit.Hash)
			out = append(out, commit)
		}
		return out, nil
	}, 0)
}
