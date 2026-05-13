package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListRepoWatchers returns all users watching a repository.
// Cloud endpoint: GET /repositories/{ws}/{slug}/watchers
func (c *Client) ListRepoWatchers(ns, slug string) ([]backend.User, error) {
	path := fmt.Sprintf("/repositories/%s/%s/watchers?pagelen=100", ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.User, error) {
		var page cloudPagedResponse[wireCloudUser]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.User, 0, len(page.Values))
		for i := range page.Values {
			out = append(out, page.Values[i].toDomain())
		}
		return out, nil
	}, 0)
}
