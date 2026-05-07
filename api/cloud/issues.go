package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireCloudUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (w *wireCloudUser) toDomain() backend.User {
	if w == nil {
		return backend.User{}
	}
	return backend.User{Slug: w.Username, DisplayName: w.DisplayName}
}

type wireCloudIssue struct {
	ID        int            `json:"id"`
	Title     string         `json:"title"`
	State     string         `json:"state"`
	Kind      string         `json:"kind"`
	Priority  string         `json:"priority"`
	Reporter  *wireCloudUser `json:"reporter"`
	Assignee  *wireCloudUser `json:"assignee"`
	CreatedOn string         `json:"created_on"`
	UpdatedOn string         `json:"updated_on"`
	Links     struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
}

func (w wireCloudIssue) toDomain() backend.Issue {
	issue := backend.Issue{
		ID:       w.ID,
		Title:    w.Title,
		State:    w.State,
		Kind:     w.Kind,
		Priority: w.Priority,
		Reporter: w.Reporter.toDomain(),
		WebURL:   w.Links.HTML.Href,
		Content:  w.Content.Raw,
	}
	if w.Assignee != nil {
		u := w.Assignee.toDomain()
		issue.Assignee = &u
	}
	if t, err := time.Parse(time.RFC3339, w.CreatedOn); err == nil {
		issue.CreatedOn = t
	}
	if t, err := time.Parse(time.RFC3339, w.UpdatedOn); err == nil {
		issue.UpdatedOn = t
	}
	return issue
}

// ListIssues fetches issues from a repository's tracker. The state argument
// is appended as a `q=state="..."` filter when non-empty. Bitbucket Cloud's
// query language is case-sensitive on field names but accepts lowercase
// state values exactly.
func (c *Client) ListIssues(ns, slug, state string, limit int) ([]backend.Issue, error) {
	if ns == "" || slug == "" {
		return nil, fmt.Errorf("workspace and repo required")
	}
	q := url.Values{}
	if state != "" {
		q.Set("q", fmt.Sprintf(`state="%s"`, state))
	}
	if limit > 0 {
		q.Set("pagelen", fmt.Sprintf("%d", limit))
	}
	q.Set("sort", "-created_on")
	path := fmt.Sprintf("/repositories/%s/%s/issues?%s", ns, slug, q.Encode())

	return paging.Collect(c.http, path, func(body []byte) ([]backend.Issue, error) {
		var page cloudPagedResponse[wireCloudIssue]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Issue, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

func (c *Client) GetIssue(ns, slug string, id int) (backend.Issue, error) {
	var w wireCloudIssue
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", ns, slug, id)
	if err := c.getJSON(path, &w); err != nil {
		return backend.Issue{}, err
	}
	return w.toDomain(), nil
}

// createIssueBody mirrors the subset of fields Bitbucket Cloud accepts on POST.
// We use a struct (not map) to give json.Marshal a stable key order — the
// server doesn't care, but stable JSON makes test assertions tractable.
type createIssueBody struct {
	Title    string            `json:"title"`
	Content  *issueContentBody `json:"content,omitempty"`
	Kind     string            `json:"kind,omitempty"`
	Priority string            `json:"priority,omitempty"`
}

type issueContentBody struct {
	Raw string `json:"raw"`
}

func (c *Client) CreateIssue(ns, slug string, in backend.CreateIssueInput) (backend.Issue, error) {
	if in.Title == "" {
		return backend.Issue{}, fmt.Errorf("title required")
	}
	body := createIssueBody{
		Title:    in.Title,
		Kind:     in.Kind,
		Priority: in.Priority,
	}
	if in.Content != "" {
		body.Content = &issueContentBody{Raw: in.Content}
	}
	var w wireCloudIssue
	path := fmt.Sprintf("/repositories/%s/%s/issues", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Issue{}, err
	}
	return w.toDomain(), nil
}

// updateIssueBody — same all-pointer rationale as the create body.
type updateIssueBody struct {
	Title    string            `json:"title,omitempty"`
	State    string            `json:"state,omitempty"`
	Kind     string            `json:"kind,omitempty"`
	Priority string            `json:"priority,omitempty"`
	Content  *issueContentBody `json:"content,omitempty"`
}

func (c *Client) UpdateIssue(ns, slug string, id int, in backend.UpdateIssueInput) (backend.Issue, error) {
	body := updateIssueBody{
		Title:    in.Title,
		State:    in.State,
		Kind:     in.Kind,
		Priority: in.Priority,
	}
	if in.Content != "" {
		body.Content = &issueContentBody{Raw: in.Content}
	}
	var w wireCloudIssue
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", ns, slug, id)
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.Issue{}, err
	}
	return w.toDomain(), nil
}
