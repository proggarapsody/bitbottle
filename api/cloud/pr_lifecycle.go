package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// UpdatePR updates the title and/or description of a pull request.
func (c *Client) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	body := map[string]string{
		"title":       in.Title,
		"description": in.Description,
	}
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

// DeclinePR declines an open pull request.
// A nil body is intentional: ContentTypeWhenBody ensures no Content-Type is
// set, which is required for this endpoint on Bitbucket Cloud.
func (c *Client) DeclinePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/decline", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

// UnapprovePR removes the authenticated user's approval from a pull request.
func (c *Client) UnapprovePR(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve", ns, slug, id)
	return c.delete(path)
}

// ReadyPR marks a draft pull request as ready for review.
func (c *Client) ReadyPR(ns, slug string, id int) error {
	body := map[string]bool{"draft": false}
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)
	return c.putJSON(path, body, &result)
}

// RequestReview adds reviewers to a pull request using PUT /pullrequests/{id}.
// It first GETs the current PR to preserve existing reviewers, then PUTs the
// merged list. This is the only Cloud-supported approach (the /participants
// endpoint does not support adding reviewers).
func (c *Client) RequestReview(ns, slug string, id int, users []string) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)

	// 1. Read current PR to preserve existing reviewers.
	var current wireCloudPR
	if err := c.getJSON(path, &current); err != nil {
		return err
	}

	// 2. Merge existing + new reviewers, deduplicated.
	seen := make(map[string]bool)
	type reviewer struct {
		AccountID string `json:"account_id"`
	}
	var reviewers []reviewer
	for _, r := range current.Reviewers {
		if !seen[r.AccountID] {
			seen[r.AccountID] = true
			reviewers = append(reviewers, reviewer{AccountID: r.AccountID})
		}
	}
	for _, u := range users {
		if !seen[u] {
			seen[u] = true
			reviewers = append(reviewers, reviewer{AccountID: u})
		}
	}

	// 3. PUT with updated reviewers (title is required by the Cloud API).
	body := map[string]any{
		"title":     current.Title,
		"reviewers": reviewers,
	}
	var result wireCloudPR
	return c.putJSON(path, body, &result)
}

// RequestChangesPR requests changes on a pull request (Cloud only).
// A nil body is intentional: ContentTypeWhenBody ensures no Content-Type is
// set, which is required for this endpoint on Bitbucket Cloud.
func (c *Client) RequestChangesPR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/request-changes", ns, slug, id)
	return c.postJSON(path, nil, &result)
}
