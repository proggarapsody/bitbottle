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
// Assignee uses a pointer-to-struct so we can send null to clear the assignee.
type updateIssueBody struct {
	Title    string             `json:"title,omitempty"`
	State    string             `json:"state,omitempty"`
	Kind     string             `json:"kind,omitempty"`
	Priority string             `json:"priority,omitempty"`
	Content  *issueContentBody  `json:"content,omitempty"`
	Assignee *issueAssigneeBody `json:"assignee,omitempty"`
}

// issueAssigneeBody is the wire shape for setting/clearing the assignee.
// When Username is empty and the pointer is non-nil, Cloud interprets
// the object as "unassign" — we only use it when explicitly requested.
type issueAssigneeBody struct {
	Username string `json:"username,omitempty"`
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
	if in.Assignee != "" {
		if in.Assignee == backend.AssigneeNone {
			body.Assignee = &issueAssigneeBody{} // empty object = unassign
		} else {
			body.Assignee = &issueAssigneeBody{Username: in.Assignee}
		}
	}
	var w wireCloudIssue
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", ns, slug, id)
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.Issue{}, err
	}
	return w.toDomain(), nil
}

// ReopenIssue transitions an issue back to the "open" state.
func (c *Client) ReopenIssue(ns, slug string, id int) error {
	_, err := c.UpdateIssue(ns, slug, id, backend.UpdateIssueInput{State: "open"})
	return err
}

// AssignIssue sets the assignee on an issue by username.
func (c *Client) AssignIssue(ns, slug string, id int, assignee string) error {
	_, err := c.UpdateIssue(ns, slug, id, backend.UpdateIssueInput{Assignee: assignee})
	return err
}

// wireCloudIssueComment is the Bitbucket Cloud wire shape for an issue comment.
type wireCloudIssueComment struct {
	ID      int            `json:"id"`
	Author  *wireCloudUser `json:"author"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	CreatedOn string `json:"created_on"`
	UpdatedOn string `json:"updated_on"`
}

func (w wireCloudIssueComment) toDomain() backend.IssueComment {
	c := backend.IssueComment{
		ID:      w.ID,
		Content: w.Content.Raw,
	}
	if w.Author != nil {
		c.Author = w.Author.toDomain()
	}
	if t, err := time.Parse(time.RFC3339, w.CreatedOn); err == nil {
		c.CreatedOn = t
	}
	if t, err := time.Parse(time.RFC3339, w.UpdatedOn); err == nil {
		c.UpdatedOn = t
	}
	return c
}

// issueCommentBody is the request body for creating/editing an issue comment.
type issueCommentBody struct {
	Content issueContentBody `json:"content"`
}

func (c *Client) ListIssueComments(ns, slug string, id int) ([]backend.IssueComment, error) {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments", ns, slug, id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.IssueComment, error) {
		var page cloudPagedResponse[wireCloudIssueComment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.IssueComment, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

func (c *Client) AddIssueComment(ns, slug string, id int, body string) (backend.IssueComment, error) {
	reqBody := issueCommentBody{Content: issueContentBody{Raw: body}}
	var w wireCloudIssueComment
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments", ns, slug, id)
	if err := c.postJSON(path, reqBody, &w); err != nil {
		return backend.IssueComment{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) EditIssueComment(ns, slug string, id, commentID int, body string) (backend.IssueComment, error) {
	reqBody := issueCommentBody{Content: issueContentBody{Raw: body}}
	var w wireCloudIssueComment
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments/%d", ns, slug, id, commentID)
	if err := c.putJSON(path, reqBody, &w); err != nil {
		return backend.IssueComment{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteIssueComment(ns, slug string, id, commentID int) error {
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments/%d", ns, slug, id, commentID)
	return c.delete(path)
}
