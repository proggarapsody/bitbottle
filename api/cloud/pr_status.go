package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListMyPRs returns the authenticated user's open PRs from the Cloud
// dashboard endpoint, combining AUTHOR and REVIEWER roles.
// Duplicates (PR appearing in both roles) keep the AUTHOR entry.
func (c *Client) ListMyPRs(ns, slug string) ([]backend.MyPREntry, error) {
	seen := make(map[int]backend.MyPREntry)

	for _, role := range []string{"REVIEWER", "AUTHOR"} {
		path := fmt.Sprintf("/dashboard/pullrequests?role=%s&state=OPEN&pagelen=50", role)
		entries, err := paging.Collect(c.http, path, func(body []byte) ([]backend.MyPREntry, error) {
			var page cloudPagedResponse[cloudgen.CloudDashboardPR]
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.MyPREntry, 0, len(page.Values))
			for _, w := range page.Values {
				entry := backend.MyPREntry{
					PullRequest: backend.PullRequest{
						ID:    w.ID,
						Title: w.Title,
						State: w.State,
						Author: backend.User{
							Slug:        w.Author.Nickname,
							DisplayName: w.Author.DisplayName,
						},
						WebURL:         w.Links.HTML.Href,
						HeadCommitHash: w.Source.Commit.Hash,
					},
					Repo: w.Source.Repository.FullName,
					Role: role,
				}
				out = append(out, entry)
			}
			return out, nil
		}, 0)
		if err != nil {
			return nil, err
		}
		// AUTHOR wins on conflict (iterate REVIEWER first, then AUTHOR overwrites)
		for _, e := range entries {
			seen[e.ID] = e
		}
	}

	result := make([]backend.MyPREntry, 0, len(seen))
	for _, e := range seen {
		result = append(result, e)
	}
	return result, nil
}
