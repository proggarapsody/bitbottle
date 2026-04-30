package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (c *Client) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	body := map[string]string{
		"title":       in.Title,
		"description": in.Description,
	}
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

// DeclinePR declines an open pull request.
func (c *Client) DeclinePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/decline", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

// UnapprovePR removes the authenticated user's approval from a pull request.
// Mirrors the approve endpoint: DELETE .../approve (not DELETE .../participants/~,
// which requires an actual user slug and is rejected by Bitbucket Server).
func (c *Client) UnapprovePR(ns, slug string, id int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/approve", ns, slug, id)
	return c.delete(path, nil)
}

// ReadyPR marks a draft pull request as ready for review.
//
// Bitbucket Server's PUT endpoint for a PR requires the full PR object
// (title, fromRef, toRef, ...), so we GET the current PR first, flip the
// draft flag, and PUT the full body back.
func (c *Client) ReadyPR(ns, slug string, id int) error {
	var current wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &current); err != nil {
		return err
	}
	current.Draft = false
	var result struct{}
	return c.putJSON(path, current, &result)
}

// wireReviewer is the wire type for a reviewer entry in the Server PR body.
type wireReviewer struct {
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

// wireReviewerPR is the body used when PUTting reviewers back onto a PR.
type wireReviewerPR struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Reviewers   []wireReviewer `json:"reviewers"`
}

// wirePRWithReviewers extends wirePR to capture the existing reviewers list.
type wirePRWithReviewers struct {
	wirePR
	Reviewers []wireReviewer `json:"reviewers"`
}

func (c *Client) RequestReview(ns, slug string, id int, users []string) error {
	var current wirePRWithReviewers
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &current); err != nil {
		return err
	}

	existing := make(map[string]struct{}, len(current.Reviewers))
	merged := make([]wireReviewer, 0, len(current.Reviewers)+len(users))
	for _, r := range current.Reviewers {
		existing[r.User.Name] = struct{}{}
		merged = append(merged, r)
	}
	for _, u := range users {
		if _, ok := existing[u]; !ok {
			var r wireReviewer
			r.User.Name = u
			merged = append(merged, r)
		}
	}

	body := wireReviewerPR{
		Title:       current.Title,
		Description: current.Description,
		Reviewers:   merged,
	}
	var result struct{}
	return c.putJSON(path, body, &result)
}
