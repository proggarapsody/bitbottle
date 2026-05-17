package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

// cloudInlineToDomain maps Bitbucket Cloud's inline payload to the domain type.
// Returns nil if the inline anchor has no usable line number.
func cloudInlineToDomain(in *cloudgen.CloudInline) *backend.PRCommentInline {
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

func toPRCommentDomain(w cloudgen.CloudPRComment) backend.PRComment {
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
		var page cloudPagedResponse[cloudgen.CloudPRComment]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, w := range page.Values {
			out = append(out, toPRCommentDomain(w))
		}
		return nil
	})
	return out, err
}

// inlineDomainToCloud maps the domain inline anchor onto Cloud's wire shape.
// Side="new" populates `to` (and `start_to` for multi-line); Side="old"
// populates `from` (and `start_from`).
func inlineDomainToCloud(in *backend.PRCommentInline) *cloudgen.CloudInline {
	if in == nil {
		return nil
	}
	out := &cloudgen.CloudInline{Path: in.Path}
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
	in := cloudgen.CloudAddPRComment{
		Content: cloudgen.CloudPRCommentContent{Raw: body},
	}
	var w cloudgen.CloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", ns, slug, id, commentID)
	if err := c.putJSON(path, in, &w); err != nil {
		return backend.PRComment{}, err
	}
	return toPRCommentDomain(w), nil
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
	var w cloudgen.CloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", ns, slug, id, commentID)
	return c.putJSON(path, body, &w)
}

func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := cloudgen.CloudAddPRComment{
		Content: cloudgen.CloudPRCommentContent{Raw: in.Text},
		Inline:  inlineDomainToCloud(in.Inline),
	}
	if in.Parent != nil {
		body.Parent = &cloudgen.CloudParentRef{ID: *in.Parent}
	}

	var w cloudgen.CloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return toPRCommentDomain(w), nil
}
