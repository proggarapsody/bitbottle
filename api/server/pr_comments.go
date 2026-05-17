package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func prCommentBaseDomain(w servergen.RestServerPRComment, parentID int, inline *backend.PRCommentInline) backend.PRComment {
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
		Severity:  w.Severity,
		State:     w.State,
		Version:   w.Version,
	}
	if w.UpdatedDate != 0 {
		c.UpdatedAt = time.UnixMilli(w.UpdatedDate).UTC()
	}
	return c
}

func toPRCommentDomain(w servergen.RestServerPRComment) backend.PRComment {
	return prCommentBaseDomain(w, 0, nil)
}

// serverAnchorToInline maps Bitbucket Server's commentAnchor envelope to the
// domain inline type. Returns nil if the anchor has no path or line.
func serverAnchorToInline(a *servergen.RestCommentAnchor) *backend.PRCommentInline {
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
func flattenServerReplies(top servergen.RestServerPRComment, inline *backend.PRCommentInline, parentID int) []backend.PRComment {
	out := []backend.PRComment{prCommentBaseDomain(top, parentID, inline)}
	for _, child := range top.Comments {
		out = append(out, flattenServerReplies(child, nil, top.ID)...)
	}
	return out
}

func (c *Client) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/activities?limit=100", ns, slug, id)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PRComment, error) {
		var page PagedResponse[servergen.RestPRActivity]
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

// fetchPRDiffHashes returns the (fromHash, toHash) for the given path on a
// pull request. Server requires both on every inline-comment anchor; the
// JSON-flavoured /diff/{path} endpoint surfaces them at the envelope root.
func (c *Client) fetchPRDiffHashes(ns, slug string, id int, filePath string) (string, string, error) {
	var env servergen.RestDiffEnvelope
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/diff/%s", ns, slug, id, filePath)
	if err := c.getJSON(path, &env); err != nil {
		return "", "", err
	}
	return env.FromHash, env.ToHash, nil
}

func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := servergen.RestAddPRComment{Text: in.Text}
	if in.Severity != "" {
		body.Severity = &in.Severity
	}

	if in.Inline != nil {
		if in.Inline.StartLine != 0 && in.Inline.StartLine != in.Inline.Line {
			return backend.PRComment{}, fmt.Errorf("multi-line inline comments are not supported on Bitbucket Server / Data Center; use a single-line --inline path:line")
		}
		fromHash, toHash, err := c.fetchPRDiffHashes(ns, slug, id, in.Inline.Path)
		if err != nil {
			return backend.PRComment{}, err
		}
		fileType := "TO"
		lineType := "ADDED"
		if in.Inline.Side == "old" {
			fileType = "FROM"
			lineType = "REMOVED"
		}
		body.Anchor = &servergen.RestCommentAnchor{
			Path:     in.Inline.Path,
			Line:     in.Inline.Line,
			LineType: lineType,
			FileType: fileType,
			FromHash: fromHash,
			ToHash:   toHash,
			SrcPath:  in.Inline.Path,
		}
	}
	if in.Parent != nil {
		body.Parent = &servergen.RestParentRef{ID: *in.Parent}
	}

	var w servergen.RestServerPRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return toPRCommentDomain(w), nil
}

// fetchCommentVersion looks up the current version of a comment so the
// caller can pass it to a subsequent edit/delete. Server's optimistic
// locking demands an exact match; we fetch fresh on each write rather
// than maintain a cross-call cache.
func (c *Client) fetchCommentVersion(ns, slug string, id, commentID int) (int, error) {
	var v struct {
		Version int `json:"version"`
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d", ns, slug, id, commentID)
	if err := c.getJSON(path, &v); err != nil {
		return 0, err
	}
	return v.Version, nil
}

func (c *Client) EditPRComment(ns, slug string, id, commentID int, body string) (backend.PRComment, error) {
	version, err := c.fetchCommentVersion(ns, slug, id, commentID)
	if err != nil {
		return backend.PRComment{}, err
	}
	in := servergen.RestEditPRComment{Text: body, Version: version}
	var w servergen.RestServerPRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d?version=%d", ns, slug, id, commentID, version)
	if err := c.putJSON(path, in, &w); err != nil {
		return backend.PRComment{}, err
	}
	return toPRCommentDomain(w), nil
}

func (c *Client) DeletePRComment(ns, slug string, id, commentID int) error {
	version, err := c.fetchCommentVersion(ns, slug, id, commentID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d?version=%d", ns, slug, id, commentID, version)
	return c.delete(path, nil)
}

// SetPRCommentState sets the state of a task comment (BLOCKER severity) on a
// pull request. It first fetches the current comment to get the version token
// (GET), then issues a PUT with the new state and the fetched version.
// state must be "OPEN" or "RESOLVED".
func (c *Client) SetPRCommentState(ns, slug string, id, commentID int, state string) error {
	version, err := c.fetchCommentVersion(ns, slug, id, commentID)
	if err != nil {
		return err
	}
	in := servergen.RestSetCommentState{State: state, Version: version}
	var w servergen.RestServerPRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments/%d?version=%d", ns, slug, id, commentID, version)
	return c.putJSON(path, in, &w)
}
