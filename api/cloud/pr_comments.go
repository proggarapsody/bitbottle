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
