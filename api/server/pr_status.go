package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireInboxPR is the shape of items from the Server inbox pull-requests endpoint
// (PRs assigned to the user for review).
type wireInboxPR struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	State string `json:"state"`
	ToRef struct {
		Repository struct {
			Project struct {
				Key string `json:"key"`
			} `json:"project"`
			Slug string `json:"slug"`
		} `json:"repository"`
	} `json:"toRef"`
	FromRef struct {
		LatestCommit string `json:"latestCommit"`
	} `json:"fromRef"`
	Author struct {
		User struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"author"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

// ListMyPRs returns the authenticated user's open PRs.
// REVIEWER PRs come from /inbox/pull-requests.
// AUTHOR PRs come from the scoped repo endpoint filtered by author slug.
func (c *Client) ListMyPRs(ns, slug string) ([]backend.MyPREntry, error) {
	seen := make(map[int]backend.MyPREntry)

	// Fetch REVIEWER PRs from the inbox endpoint
	reviewerEntries, err := c.listInboxPRs()
	if err != nil {
		return nil, err
	}
	for _, e := range reviewerEntries {
		seen[e.ID] = e
	}

	// Fetch AUTHOR PRs from the scoped repo endpoint
	currentUser, err := c.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	authorEntries, err := c.listAuthorPRs(ns, slug, currentUser.Slug)
	if err != nil {
		return nil, err
	}
	// AUTHOR wins on conflict
	for _, e := range authorEntries {
		seen[e.ID] = e
	}

	result := make([]backend.MyPREntry, 0, len(seen))
	for _, e := range seen {
		result = append(result, e)
	}
	return result, nil
}

func (c *Client) listInboxPRs() ([]backend.MyPREntry, error) {
	path := "/inbox/pull-requests?start=0&limit=50"
	body, err := c.getBytes(path)
	if err != nil {
		return nil, err
	}
	var page PagedResponse[wireInboxPR]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}
	out := make([]backend.MyPREntry, 0, len(page.Values))
	for _, w := range page.Values {
		webURL := ""
		if len(w.Links.Self) > 0 {
			webURL = w.Links.Self[0].Href
		}
		key := w.ToRef.Repository.Project.Key
		repoSlug := w.ToRef.Repository.Slug
		entry := backend.MyPREntry{
			PullRequest: backend.PullRequest{
				ID:    w.ID,
				Title: w.Title,
				State: w.State,
				Author: backend.User{
					Slug:        w.Author.User.Slug,
					DisplayName: w.Author.User.DisplayName,
				},
				WebURL:         webURL,
				HeadCommitHash: w.FromRef.LatestCommit,
			},
			Repo: fmt.Sprintf("%s/%s", key, repoSlug),
			Role: "REVIEWER",
		}
		out = append(out, entry)
	}
	return out, nil
}

func (c *Client) listAuthorPRs(ns, slug, userSlug string) ([]backend.MyPREntry, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests?author=%s&state=OPEN&start=0&limit=50", ns, slug, userSlug)
	body, err := c.getBytes(path)
	if err != nil {
		return nil, err
	}
	var page PagedResponse[wirePR]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}
	out := make([]backend.MyPREntry, 0, len(page.Values))
	for _, w := range page.Values {
		pr := w.toDomain()
		entry := backend.MyPREntry{
			PullRequest: pr,
			Repo:        fmt.Sprintf("%s/%s", ns, slug),
			Role:        "AUTHOR",
		}
		out = append(out, entry)
	}
	return out, nil
}
