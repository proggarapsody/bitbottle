package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toCommitCommentDomain(w servergen.RestCommitComment) backend.CommitComment {
	c := backend.CommitComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        w.Author.Slug,
			DisplayName: w.Author.DisplayName,
		},
		Body:      w.Text,
		CreatedAt: time.UnixMilli(w.CreatedDate).UTC(),
	}
	if w.UpdatedDate != 0 {
		c.UpdatedAt = time.UnixMilli(w.UpdatedDate).UTC()
	}
	return c
}

// ListCommitComments lists all comments on a commit using the standard
// Bitbucket Server paged response. Uses paging.Collect to handle pagination.
func (c *Client) ListCommitComments(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments?limit=100", ns, slug, hash)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.CommitComment, error) {
		var page PagedResponse[servergen.RestCommitComment]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.CommitComment, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toCommitCommentDomain(w))
		}
		return out, nil
	}, limit)
}

// AddCommitComment posts a new comment on a commit.
func (c *Client) AddCommitComment(ns, slug, hash string, in backend.AddCommitCommentInput) (backend.CommitComment, error) {
	body := servergen.RestAddCommitComment{Text: in.Body}
	var w servergen.RestCommitComment
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments", ns, slug, hash)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.CommitComment{}, err
	}
	return toCommitCommentDomain(w), nil
}

// fetchCommitCommentVersion looks up the current version of a commit comment
// so the caller can pass it to a subsequent edit/delete. Server's optimistic
// locking requires an exact match; we fetch fresh on each write.
func (c *Client) fetchCommitCommentVersion(ns, slug, hash string, commentID int) (int, error) {
	var v struct {
		Version int `json:"version"`
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d", ns, slug, hash, commentID)
	if err := c.getJSON(path, &v); err != nil {
		return 0, err
	}
	return v.Version, nil
}

// EditCommitComment updates the body of a commit comment. Server requires
// the current version for optimistic concurrency — the comment is fetched
// first to obtain it, then the PUT is issued.
func (c *Client) EditCommitComment(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error) {
	version, err := c.fetchCommitCommentVersion(ns, slug, hash, commentID)
	if err != nil {
		return backend.CommitComment{}, err
	}
	in := servergen.RestEditCommitComment{Text: body, Version: version}
	var w servergen.RestCommitComment
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d?version=%d", ns, slug, hash, commentID, version)
	if err := c.putJSON(path, in, &w); err != nil {
		return backend.CommitComment{}, err
	}
	return toCommitCommentDomain(w), nil
}

// DeleteCommitComment removes a commit comment. Server requires the current
// version as a query param for optimistic concurrency.
func (c *Client) DeleteCommitComment(ns, slug, hash string, commentID int) error {
	version, err := c.fetchCommitCommentVersion(ns, slug, hash, commentID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/commits/%s/comments/%d?version=%d", ns, slug, hash, commentID, version)
	return c.delete(path, nil)
}
