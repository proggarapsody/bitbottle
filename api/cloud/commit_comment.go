package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toCommitCommentDomain(w cloudgen.CloudCommitComment) backend.CommitComment {
	slug := w.User.Nickname
	if slug == "" {
		slug = w.User.AccountID
	}
	return backend.CommitComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        slug,
			DisplayName: w.User.DisplayName,
		},
		Body:      w.Content.Raw,
		CreatedAt: w.CreatedOn,
		UpdatedAt: w.UpdatedOn,
	}
}

// ListCommitComments lists all comments on a commit. Cloud supports
// pagination via the standard paged response envelope.
func (c *Client) ListCommitComments(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
	path := fmt.Sprintf("/repositories/%s/%s/commits/%s/comments?pagelen=100", ns, slug, hash)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.CommitComment, error) {
		var page cloudPagedResponse[cloudgen.CloudCommitComment]
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
	body := cloudgen.CloudAddCommitComment{
		Content: cloudgen.CloudCommitCommentContent{Raw: in.Body},
	}
	var w cloudgen.CloudCommitComment
	path := fmt.Sprintf("/repositories/%s/%s/commits/%s/comments", ns, slug, hash)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.CommitComment{}, err
	}
	return toCommitCommentDomain(w), nil
}

// EditCommitComment updates the body of an existing commit comment via PUT.
func (c *Client) EditCommitComment(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error) {
	req := cloudgen.CloudAddCommitComment{
		Content: cloudgen.CloudCommitCommentContent{Raw: body},
	}
	var w cloudgen.CloudCommitComment
	path := fmt.Sprintf("/repositories/%s/%s/commits/%s/comments/%d", ns, slug, hash, commentID)
	if err := c.putJSON(path, req, &w); err != nil {
		return backend.CommitComment{}, err
	}
	return toCommitCommentDomain(w), nil
}

// DeleteCommitComment removes a commit comment. Cloud returns 204 on success.
func (c *Client) DeleteCommitComment(ns, slug, hash string, commentID int) error {
	path := fmt.Sprintf("/repositories/%s/%s/commits/%s/comments/%d", ns, slug, hash, commentID)
	return c.delete(path)
}
