package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toWatcherDomain(w servergen.RestWatcher) backend.User {
	return backend.User{Slug: w.Slug, DisplayName: w.DisplayName}
}

// ListRepoWatchers returns all users watching a repository.
// Server endpoint: GET /rest/api/1.0/projects/{ns}/repos/{slug}/watchers
func (c *Client) ListRepoWatchers(ns, slug string) ([]backend.User, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/watchers", ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.User, error) {
		var page PagedResponse[servergen.RestWatcher]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.User, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toWatcherDomain(w))
		}
		return out, nil
	}, 0)
}
