package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireCloudDashboardPR is the shape of each item from the Cloud dashboard
// pull-requests endpoint.
type wireCloudDashboardPR struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Source struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"source"`
	Author struct {
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	} `json:"author"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

// ListMyPRs returns the authenticated user's open PRs from the Cloud
// dashboard endpoint, combining AUTHOR and REVIEWER roles.
// Duplicates (PR appearing in both roles) keep the AUTHOR entry.
func (c *Client) ListMyPRs(ns, slug string) ([]backend.MyPREntry, error) {
	seen := make(map[int]backend.MyPREntry)

	for _, role := range []string{"REVIEWER", "AUTHOR"} {
		path := fmt.Sprintf("/dashboard/pullrequests?role=%s&state=OPEN&pagelen=50", role)
		entries, err := paging.Collect(c.http, path, func(body []byte) ([]backend.MyPREntry, error) {
			var page cloudPagedResponse[wireCloudDashboardPR]
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
