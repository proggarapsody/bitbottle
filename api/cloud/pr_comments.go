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
}

func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := wireCloudAddPRComment{}
	body.Content.Raw = in.Text

	var w wireCloudPRComment
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return w.toDomain(), nil
}
