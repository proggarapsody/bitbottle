package server

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wirePR struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Author      struct {
		User struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"author"`
	FromRef struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
	} `json:"fromRef"`
	ToRef struct {
		ID        string `json:"id"`
		DisplayID string `json:"displayId"`
	} `json:"toRef"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

func (w wirePR) toDomain() backend.PullRequest {
	webURL := ""
	if len(w.Links.Self) > 0 {
		webURL = w.Links.Self[0].Href
	}
	return backend.PullRequest{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		State:       w.State,
		Draft:       w.Draft,
		Author: backend.User{
			Slug:        w.Author.User.Slug,
			DisplayName: w.Author.User.DisplayName,
		},
		FromBranch: w.FromRef.DisplayID,
		ToBranch:   w.ToRef.DisplayID,
		WebURL:     webURL,
	}
}

// ensureRefsHeads prepends "refs/heads/" to branch if not already present.
func ensureRefsHeads(branch string) string {
	if strings.HasPrefix(branch, "refs/heads/") {
		return branch
	}
	return "refs/heads/" + branch
}

// ListPRs lists pull requests for a repository.
func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	var page PagedResponse[wirePR]
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests?state=%s&limit=%d", ns, slug, state, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	prs := make([]backend.PullRequest, 0, len(page.Values))
	for _, w := range page.Values {
		prs = append(prs, w.toDomain())
	}
	return prs, nil
}

// GetPR fetches a single pull request.
func (c *Client) GetPR(ns, slug string, id int) (backend.PullRequest, error) {
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

type wireCreatePRInput struct {
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Draft       bool        `json:"draft,omitempty"`
	FromRef     wireRefBody `json:"fromRef"`
	ToRef       wireRefBody `json:"toRef"`
}

type wireRefBody struct {
	ID string `json:"id"`
}

// CreatePR creates a new pull request.
func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := wireCreatePRInput{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		FromRef:     wireRefBody{ID: ensureRefsHeads(in.FromBranch)},
		ToRef:       wireRefBody{ID: ensureRefsHeads(in.ToBranch)},
	}
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

type wireMergePRInput struct {
	Message  string `json:"message,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// MergePR merges a pull request.
func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	body := wireMergePRInput{
		Message:  in.Message,
		Strategy: in.Strategy,
	}
	var w wirePR
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/merge", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
func (c *Client) ApprovePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/participants/~", ns, slug, id)
	return c.putJSON(path, map[string]string{"status": "APPROVED"}, &result)
}

// GetPRDiff fetches the unified diff for a PR.
func (c *Client) GetPRDiff(ns, slug string, id int) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/diff", ns, slug, id)
	return c.getText(path)
}

// DeleteBranch deletes a branch in a repository.
func (c *Client) DeleteBranch(ns, slug, branch string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/branches", ns, slug)
	return c.delete(path, map[string]string{"name": branch})
}

// GetCurrentUser fetches the authenticated user.
// Bitbucket Server does not support GET /users/~ (Cloud-only), so when a
// userSlug was provided at construction time we call GET /users/{slug} instead.
func (c *Client) GetCurrentUser() (backend.User, error) {
	path := "/users/~"
	if c.userSlug != "" {
		path = "/users/" + c.userSlug
	}
	var w struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	}
	if err := c.getJSON(path, &w); err != nil {
		return backend.User{}, err
	}
	return backend.User{Slug: w.Slug, DisplayName: w.DisplayName}, nil
}
