package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudPR struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Draft       bool   `json:"draft"`
	Author      struct {
		DisplayName string `json:"display_name"`
		AccountID   string `json:"account_id"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
	Links struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

func (w wireCloudPR) toDomain() backend.PullRequest {
	return backend.PullRequest{
		ID:          w.ID,
		Title:       w.Title,
		Description: w.Description,
		State:       w.State,
		Draft:       w.Draft,
		Author: backend.User{
			Slug:        w.Author.AccountID,
			DisplayName: w.Author.DisplayName,
		},
		FromBranch: w.Source.Branch.Name,
		ToBranch:   w.Destination.Branch.Name,
		WebURL:     w.Links.HTML.Href,
	}
}

// ListPRs lists pull requests for a repository.
func (c *Client) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	var page cloudPagedResponse[wireCloudPR]
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=%d", ns, slug, state, limit)
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
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

type wireCloudCreatePR struct {
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Draft       bool               `json:"draft,omitempty"`
	Source      wireCloudBranchRef `json:"source"`
	Destination wireCloudBranchRef `json:"destination"`
}

type wireCloudBranchRef struct {
	Branch wireCloudBranchName `json:"branch"`
}

type wireCloudBranchName struct {
	Name string `json:"name"`
}

// CreatePR creates a new pull request.
func (c *Client) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	body := wireCloudCreatePR{
		Title:       in.Title,
		Description: in.Description,
		Draft:       in.Draft,
		Source:      wireCloudBranchRef{Branch: wireCloudBranchName{Name: in.FromBranch}},
		Destination: wireCloudBranchRef{Branch: wireCloudBranchName{Name: in.ToBranch}},
	}
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

type wireCloudMergePR struct {
	MergeStrategy string `json:"merge_strategy,omitempty"`
}

// MergePR merges a pull request.
func (c *Client) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	body := wireCloudMergePR{
		MergeStrategy: in.Strategy,
	}
	var w wireCloudPR
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PullRequest{}, err
	}
	return w.toDomain(), nil
}

// ApprovePR approves a PR on behalf of the authenticated user.
func (c *Client) ApprovePR(ns, slug string, id int) error {
	var result struct{}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve", ns, slug, id)
	return c.postJSON(path, nil, &result)
}

// GetPRDiff fetches the unified diff for a PR.
func (c *Client) GetPRDiff(ns, slug string, id int) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff", ns, slug, id)
	return c.getText(path)
}

// DeleteBranch deletes a branch in a repository.
func (c *Client) DeleteBranch(ns, slug, branch string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", ns, slug, url.PathEscape(branch))
	return c.delete(path)
}

// GetCurrentUser fetches the authenticated user.
func (c *Client) GetCurrentUser() (backend.User, error) {
	var w struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	}
	if err := c.getJSON("/user", &w); err != nil {
		return backend.User{}, err
	}
	return backend.User{
		Slug:        w.AccountID,
		DisplayName: w.DisplayName,
	}, nil
}
