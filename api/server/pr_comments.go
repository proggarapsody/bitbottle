package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireServerPRComment struct {
	ID     int    `json:"id"`
	Text   string `json:"text"`
	Author struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
	CreatedDate int64                 `json:"createdDate"` // Unix milliseconds
	UpdatedDate int64                 `json:"updatedDate"`
	Comments    []wireServerPRComment `json:"comments"` // nested replies
}

func (w wireServerPRComment) baseDomain(parentID int, inline *backend.PRCommentInline) backend.PRComment {
	c := backend.PRComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        w.Author.Slug,
			DisplayName: w.Author.DisplayName,
		},
		Text:      w.Text,
		CreatedAt: time.UnixMilli(w.CreatedDate).UTC(),
		ParentID:  parentID,
		Inline:    inline,
	}
	if w.UpdatedDate != 0 {
		c.UpdatedAt = time.UnixMilli(w.UpdatedDate).UTC()
	}
	return c
}

func (w wireServerPRComment) toDomain() backend.PRComment {
	return w.baseDomain(0, nil)
}

// wireServerCommentAnchor describes where an inline comment is anchored in
// the diff. lineType (CONTEXT/ADDED/REMOVED) is informational; fileType
// (FROM/TO) determines which side of the diff the comment lives on.
type wireServerCommentAnchor struct {
	Path     string `json:"path"`
	Line     int    `json:"line"`
	LineType string `json:"lineType"`
	FileType string `json:"fileType"`
	FromHash string `json:"fromHash"`
	ToHash   string `json:"toHash"`
	SrcPath  string `json:"srcPath"`
}

// serverAnchorToInline maps Bitbucket Server's commentAnchor envelope to the
// domain inline type. Returns nil if the anchor has no path or line.
func serverAnchorToInline(a *wireServerCommentAnchor) *backend.PRCommentInline {
	if a == nil || a.Path == "" || a.Line == 0 {
		return nil
	}
	side := "new"
	if a.FileType == "FROM" {
		side = "old"
	}
	return &backend.PRCommentInline{
		Path: a.Path,
		Side: side,
		Line: a.Line,
	}
}

// flattenServerReplies walks a top-level comment's nested reply tree and
// returns a flat slice. The root is emitted with the supplied inline anchor
// and parentID. Replies inherit no anchor (Server does not anchor replies)
// and carry their direct parent's ID.
func flattenServerReplies(top wireServerPRComment, inline *backend.PRCommentInline, parentID int) []backend.PRComment {
	out := []backend.PRComment{top.baseDomain(parentID, inline)}
	for _, child := range top.Comments {
		out = append(out, flattenServerReplies(child, nil, top.ID)...)
	}
	return out
}

// wireServerPRActivity wraps comment payloads in a PR activity envelope, as
// returned by GET /pull-requests/{id}/activities. commentAnchor is present
// for inline comments and nil for general PR comments.
type wireServerPRActivity struct {
	Action        string                   `json:"action"`
	Comment       wireServerPRComment      `json:"comment"`
	CommentAnchor *wireServerCommentAnchor `json:"commentAnchor,omitempty"`
}

func (c *Client) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/activities?limit=100", ns, slug, id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRComment, error) {
		var page PagedResponse[wireServerPRActivity]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PRComment, 0, len(page.Values))
		for _, a := range page.Values {
			if a.Action != "COMMENTED" || a.Comment.ID == 0 {
				continue
			}
			inline := serverAnchorToInline(a.CommentAnchor)
			out = append(out, flattenServerReplies(a.Comment, inline, 0)...)
		}
		return out, nil
	}, 0)
}

type wireServerAddPRComment struct {
	Text string `json:"text"`
}

func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := wireServerAddPRComment{Text: in.Text}
	var w wireServerPRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return w.toDomain(), nil
}
