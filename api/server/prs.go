package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wirePR struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
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

func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests?state=%s&limit=%d", ns, slug, state, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PullRequest, error) {
		var page PagedResponse[wirePR]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PullRequest, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

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
	Version  int    `json:"version"`
	Message  string `json:"message,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

// MergePR merges a pull request.
// Bitbucket Server uses optimistic concurrency: the POST body must include the
// current PR version (from GET), otherwise the server returns HTTP 409.
func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	var current wirePR
	prPath := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d", ns, slug, id)
	if err := c.getJSON(prPath, &current); err != nil {
		return backend.PullRequest{}, err
	}
	body := wireMergePRInput{
		Version:  current.Version,
		Message:  in.Message,
		Strategy: in.Strategy,
	}
	var w wirePR
	if err := c.postJSON(prPath+"/merge", body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
// Bitbucket Server (like Cloud) exposes a dedicated POST .../approve endpoint;
// the participants/{userSlug} path requires an actual slug and does not accept ~.
func (c *Client) ApprovePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/approve", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

func (c *Client) GetPRDiff(ns, slug string, id int) (string, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/diff", ns, slug, id)
	return c.getText(path)
}

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
