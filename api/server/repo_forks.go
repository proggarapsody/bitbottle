package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// ListRepoForks lists forks of a repository on Bitbucket Server / Data Center.
// limit controls the maximum number of results (0 = no cap).
func (c *Client) ListRepoForks(ns, slug string, limit int) ([]backend.Repository, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/forks?limit=50", ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Repository, error) {
		var page PagedResponse[servergen.RestRepository]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Repository, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRepositoryDomain(w))
		}
		return out, nil
	}, limit)
}
