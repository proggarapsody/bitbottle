package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListPRCommits returns the commits in a pull request, paginated.
func (c *Client) ListPRCommits(ns, slug string, prID int) ([]backend.Commit, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/commits?pagelen=100", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Commit, error) {
		var page cloudPagedResponse[cloudgen.CloudCommit]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Commit, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitDomain(w))
		}
		return out, nil
	}, 0)
}
