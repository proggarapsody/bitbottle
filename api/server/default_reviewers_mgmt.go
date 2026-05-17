package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toDefaultReviewerDomain(w servergen.RestDefaultReviewerUser) backend.DefaultReviewer {
	return backend.DefaultReviewer{
		UserSlug:     w.Slug,
		DisplayName:  w.DisplayName,
		EmailAddress: w.EmailAddress,
	}
}

// ListDefaultReviewers returns all configured default reviewers for a repository.
// GET /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/reviewers
func (c *Client) ListDefaultReviewers(ns, slug string) ([]backend.DefaultReviewer, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("project and repo required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/reviewers", ns, slug)
	return paging.Collect(c.defaultReviewersHTTP, path, func(body []byte) ([]backend.DefaultReviewer, error) {
		var page PagedResponse[servergen.RestDefaultReviewerUser]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DefaultReviewer, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toDefaultReviewerDomain(w))
		}
		return out, nil
	}, 0)
}

// AddDefaultReviewer adds a default reviewer to a repository.
// PUT /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/reviewers/{userSlug}
func (c *Client) AddDefaultReviewer(ns, slug, userSlug string) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("project and repo required")
	}
	if userSlug == "" {
		return fmt.Errorf("user required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/reviewers/%s", ns, slug, userSlug)
	return c.defaultReviewersHTTP.PutJSON(path, nil, nil)
}

// RemoveDefaultReviewer removes a default reviewer from a repository.
// DELETE /rest/default-reviewers/1.0/projects/{ns}/repos/{slug}/reviewers/{userSlug}
func (c *Client) RemoveDefaultReviewer(ns, slug, userSlug string) error {
	if ns == "" || slug == "" {
		return fmt.Errorf("project and repo required")
	}
	if userSlug == "" {
		return fmt.Errorf("user required")
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/reviewers/%s", ns, slug, userSlug)
	return c.defaultReviewersHTTP.DeleteJSON(path, nil)
}
