package cloud

import (
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
	CreatedOn time.Time `json:"created_on"`
}

func (w wireCloudPRComment) toDomain() backend.PRComment {
	slug := w.User.Nickname
	if slug == "" {
		slug = w.User.AccountID
	}
	return backend.PRComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        slug,
			DisplayName: w.User.DisplayName,
		},
		Text:      w.Content.Raw,
		CreatedAt: w.CreatedOn,
	}
}

// ListPRComments lists top-level comments on a pull request.
func (c *Client) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	var page cloudPagedResponse[wireCloudPRComment]
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments?pagelen=100", ns, slug, id)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	out := make([]backend.PRComment, 0, len(page.Values))
	for _, w := range page.Values {
		out = append(out, w.toDomain())
	}
	return out, nil
}

type wireCloudAddPRComment struct {
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
}

// AddPRComment adds a top-level comment to a pull request.
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
