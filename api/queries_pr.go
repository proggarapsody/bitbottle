package api

import (
	"fmt"
	"net/http"
)

// PullRequest represents a Bitbucket pull request.
type PullRequest struct {
	ID          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	State       string          `json:"state"`
	Draft       bool            `json:"draft"`
	Author      PRParticipant   `json:"author"`
	Reviewers   []PRParticipant `json:"reviewers"`
	FromRef     PRRef           `json:"fromRef"`
	ToRef       PRRef           `json:"toRef"`
	Links       PRLinks         `json:"links"`
}

// PRParticipant is a PR author or reviewer.
type PRParticipant struct {
	User     User   `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
}

// User is a Bitbucket user.
type User struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
}

// PRRef is a branch reference in a PR.
type PRRef struct {
	ID         string     `json:"id"`
	DisplayID  string     `json:"displayId"`
	Repository Repository `json:"repository"`
}

// PRLinks holds the PR web URL.
type PRLinks struct {
	Self []SelfLink `json:"self"`
}

// ListPRs lists pull requests for a repository.
func (c *Client) ListPRs(project, slug, state string, limit int) ([]PullRequest, error) {
	var page PagedResponse[PullRequest]
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests?state=%s&limit=%d", project, slug, state, limit)
	if err := c.GetJSON(path, &page); err != nil {
		return nil, err
	}
	return page.Values, nil
}

// GetPR fetches a single pull request.
func (c *Client) GetPR(project, slug string, id int) (PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", project, slug, id)
	if err := c.GetJSON(path, &pr); err != nil {
		return PullRequest{}, err
	}
	return pr, nil
}

// CreatePRInput is the request body for PR creation.
type CreatePRInput struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Draft       bool   `json:"draft,omitempty"`
	FromRef     PRRef  `json:"fromRef"`
	ToRef       PRRef  `json:"toRef"`
}

// CreatePR creates a new pull request.
func (c *Client) CreatePR(project, slug string, input CreatePRInput) (PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests", project, slug)
	if err := c.PostJSON(path, input, &pr); err != nil {
		return PullRequest{}, err
	}
	return pr, nil
}

// MergePRInput is the request body for PR merge.
type MergePRInput struct {
	Version  int    `json:"version"`
	Message  string `json:"message,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// MergePR merges a pull request.
func (c *Client) MergePR(project, slug string, id int, input MergePRInput) (PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/merge", project, slug, id)
	if err := c.PostJSON(path, input, &pr); err != nil {
		return PullRequest{}, err
	}
	return pr, nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
func (c *Client) ApprovePR(project, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/participants/~", project, slug, id)
	return c.PutJSON(path, map[string]string{"status": "APPROVED"}, &result)
}

// GetPRDiff fetches the unified diff for a PR.
func (c *Client) GetPRDiff(project, slug string, id int) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/diff", project, slug, id)
	return c.GetText(path)
}

// DeleteBranch deletes a branch in a repository.
func (c *Client) DeleteBranch(project, slug, branch string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/branches", project, slug)
	return c.doJSON(http.MethodDelete, path, map[string]string{"name": branch}, nil)
}

// GetCurrentUser fetches the authenticated user.
func (c *Client) GetCurrentUser() (User, error) {
	var user User
	if err := c.GetJSON("/users/~", &user); err != nil {
		return User{}, err
	}
	return user, nil
}
