package cloud

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudPRComment struct {
	ID      int `json:"id"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	User struct {
		AccountID   string `json:"account_id"`
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	} `json:"user"`
	CreatedOn  time.Time            `json:"created_on"`
	UpdatedOn  time.Time            `json:"updated_on"`
	Inline     *wireCloudInline     `json:"inline,omitempty"`
	Parent     *wireCloudParentRef  `json:"parent,omitempty"`
	Resolution *wireCloudResolution `json:"resolution,omitempty"`
}

type wireCloudInline struct {
	Path      string `json:"path"`
	From      *int   `json:"from,omitempty"`
	To        *int   `json:"to,omitempty"`
	StartFrom *int   `json:"start_from,omitempty"`
	StartTo   *int   `json:"start_to,omitempty"`
}

type wireCloudParentRef struct {
	ID int `json:"id"`
}

type wireCloudResolution struct {
	Type string `json:"type"`
}

// cloudInlineToDomain maps Bitbucket Cloud's inline payload to the domain type.
// Returns nil if the inline anchor has no usable line number.
func cloudInlineToDomain(in *wireCloudInline) *backend.PRCommentInline {
	if in == nil {
		return nil
	}
	out := &backend.PRCommentInline{Path: in.Path}
	switch {
	case in.To != nil:
		out.Side = "new"
		out.Line = *in.To
		if in.StartTo != nil {
			out.StartLine = *in.StartTo
		}
	case in.From != nil:
		out.Side = "old"
		out.Line = *in.From
		if in.StartFrom != nil {
			out.StartLine = *in.StartFrom
		}
	default:
		return nil
	}
	return out
}

func (w wireCloudPRComment) toDomain() backend.PRComment {
	slug := w.User.Nickname
	if slug == "" {
		slug = w.User.AccountID
	}
	c := backend.PRComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        slug,
			DisplayName: w.User.DisplayName,
		},
		Text:      w.Content.Raw,
		CreatedAt: w.CreatedOn,
		UpdatedAt: w.UpdatedOn,
		Inline:    cloudInlineToDomain(w.Inline),
	}
	if w.Parent != nil {
		c.ParentID = w.Parent.ID
	}
	if w.Resolution != nil && w.Resolution.Type != "" {
		c.Resolved = true
	}
	return c
}

func (c *Client) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	var out []backend.PRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments?pagelen=100", ns, slug, id)
	err := c.http.GetAllJSON(path, func(body []byte) error {
		var page cloudPagedResponse[wireCloudPRComment]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return nil
	})
	return out, err
}

type wireCloudAddPRComment struct {
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	Inline *wireCloudInline    `json:"inline,omitempty"`
	Parent *wireCloudParentRef `json:"parent,omitempty"`
}

// inlineDomainToCloud maps the domain inline anchor onto Cloud's wire shape.
// Side="new" populates `to` (and `start_to` for multi-line); Side="old"
// populates `from` (and `start_from`).
func inlineDomainToCloud(in *backend.PRCommentInline) *wireCloudInline {
	if in == nil {
		return nil
	}
	out := &wireCloudInline{Path: in.Path}
	line := in.Line
	switch in.Side {
	case "old":
		out.From = &line
		if in.StartLine != 0 && in.StartLine != in.Line {
			start := in.StartLine
			out.StartFrom = &start
		}
	default: // "new" or unset
		out.To = &line
		if in.StartLine != 0 && in.StartLine != in.Line {
			start := in.StartLine
			out.StartTo = &start
		}
	}
	return out
}

// EditPRComment updates the body of an existing comment on a pull request.
// Cloud accepts the same `{ "content": { "raw": ... } }` shape on PUT as on
// the original POST, returning the refreshed comment.
func (c *Client) EditPRComment(ns, slug string, id, commentID int, body string) (backend.PRComment, error) {
	in := wireCloudAddPRComment{}
	in.Content.Raw = body
	var w wireCloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", ns, slug, id, commentID)
	if err := c.putJSON(path, in, &w); err != nil {
		return backend.PRComment{}, err
	}
	return w.toDomain(), nil
}

// DeletePRComment removes a comment from a pull request. Cloud returns 204
// No Content on success and 404 when the comment is unknown or already gone.
func (c *Client) DeletePRComment(ns, slug string, id, commentID int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", ns, slug, id, commentID)
	return c.delete(path)
}

// ResolvePRComment marks a Cloud PR comment as resolved by writing
// `resolution.type=resolved` to the comment endpoint via PUT. Cloud has no
// dedicated resolve verb in the public REST surface; the resolution field
// of an existing comment is the documented mechanism. Server has no
// equivalent concept on regular comments and returns ErrUnsupportedOnHost
// via AsPRCommentResolver.
func (c *Client) ResolvePRComment(ns, slug string, id, commentID int) error {
	body := map[string]any{
		"resolution": map[string]string{"type": "resolved"},
	}
	var w wireCloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", ns, slug, id, commentID)
	return c.putJSON(path, body, &w)
}

func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := wireCloudAddPRComment{}
	body.Content.Raw = in.Text
	body.Inline = inlineDomainToCloud(in.Inline)
	if in.Parent != nil {
		body.Parent = &wireCloudParentRef{ID: *in.Parent}
	}

	var w wireCloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return w.toDomain(), nil
}
