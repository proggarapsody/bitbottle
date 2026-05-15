package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func (c *Client) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	body := map[string]string{
		"title":       in.Title,
		"description": in.Description,
	}
	var w servergen.RestPullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return toPRDomain(w), nil
}

// DeclinePR declines an open pull request.
func (c *Client) DeclinePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/decline", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

// ReopenPR reverses a decline, returning the PR to OPEN.
//
// Bitbucket Server uses optimistic concurrency on its PR lifecycle endpoints:
// the POST body must include the current PR version (from GET), otherwise the
// server returns HTTP 409 "Pull request was updated…" against any non-zero-
// version declined PR. Mirrors MergePR's GET-then-POST(version) pattern.
func (c *Client) ReopenPR(ns, slug string, id int) error {
	var current servergen.RestPullRequest
	prPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(prPath, &current); err != nil {
		return stampPRNotFound(err, id)
	}
	body := struct {
		Version int `json:"version"`
	}{Version: current.Version}
	var result struct{}
	return stampPRNotFound(c.postJSON(prPath+"/reopen", body, &result), id)
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
	var current servergen.RestPullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &current); err != nil {
		return err
	}
	current.Draft = false
	var result struct{}
	return c.putJSON(path, current, &result)
}

// prReviewerInput is the wire type for a reviewer entry in the Server PR body.
// Uses RestPullRequestReviewerInput from the gen package but alias kept here
// for clarity in the RequestReview method.
type prReviewerInput = servergen.RestPullRequestReviewerInput

// prWithReviewers extends RestPullRequest to capture the existing reviewers list
// when PUTting reviewers back onto a PR.
//
// Reviewers shadows the embedded field to omit omitempty: Server requires an explicit empty array to clear reviewers on PUT.
type prWithReviewers struct {
	servergen.RestPullRequest
	Reviewers []servergen.RestPullRequestReviewer `json:"reviewers"`
}

// wireReviewerPR is the body used when PUTting reviewers back onto a PR.
type wireReviewerPR struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Reviewers   []prReviewerInput `json:"reviewers"`
}

func (c *Client) RequestReview(ns, slug string, id int, users []string) error {
	var current prWithReviewers
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &current); err != nil {
		return err
	}

	existing := make(map[string]struct{}, len(current.Reviewers))
	merged := make([]prReviewerInput, 0, len(current.Reviewers)+len(users))
	for _, r := range current.Reviewers {
		existing[r.User.Name] = struct{}{}
		var ri prReviewerInput
		ri.User.Name = r.User.Name
		merged = append(merged, ri)
	}
	for _, u := range users {
		if _, ok := existing[u]; !ok {
			var r prReviewerInput
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
