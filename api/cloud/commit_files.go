package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListCommitFiles returns the files changed in a specific commit.
// Cloud endpoint: GET /repositories/{ws}/{slug}/diffstat/{hash}~1..{hash}
func (c *Client) ListCommitFiles(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
	spec := fmt.Sprintf("%s~1..%s", hash, hash)
	path := fmt.Sprintf("/repositories/%s/%s/diffstat/%s",
		url.PathEscape(ns),
		url.PathEscape(slug),
		url.PathEscape(spec),
	)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DiffStatEntry, error) {
		var page cloudPagedResponse[wireDiffStatEntry]
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
