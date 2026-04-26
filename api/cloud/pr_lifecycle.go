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

// RequestReview adds reviewers to a pull request (one request per user).
func (c *Client) RequestReview(ns, slug string, id int, users []string) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/participants", ns, slug, id)
	for _, user := range users {
		body := map[string]any{
			"user": map[string]string{
				"account_id": user,
			},
			"role": "REVIEWER",
		}
		var result struct{}
		if err := c.postJSON(path, body, &result); err != nil {
			return err
		}
	}
	return nil
}

// RequestChangesPR requests changes on a pull request (Cloud only).
func (c *Client) RequestChangesPR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/request-changes", ns, slug, id)
	return c.postJSON(path, nil, &result)
}
