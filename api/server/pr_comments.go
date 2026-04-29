package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireServerPRComment struct {
	ID     int    `json:"id"`
	Text   string `json:"text"`
	Author struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
	CreatedDate int64 `json:"createdDate"` // Unix milliseconds
}

func (w wireServerPRComment) toDomain() backend.PRComment {
	return backend.PRComment{
		ID: w.ID,
		Author: backend.User{
			Slug:        w.Author.Slug,
			DisplayName: w.Author.DisplayName,
		},
		Text:      w.Text,
		CreatedAt: time.UnixMilli(w.CreatedDate).UTC(),
	}
}

// wireServerPRActivity wraps comment payloads in a PR activity envelope, as
// returned by GET /pull-requests/{id}/activities.
type wireServerPRActivity struct {
	Action  string              `json:"action"`
	Comment wireServerPRComment `json:"comment"`
}

// ListPRComments lists top-level comments on a pull request. Bitbucket Server
// exposes comments via the activities feed; we filter for COMMENTED actions.
// Follows all pagination pages.
func (c *Client) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	var out []backend.PRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/activities?limit=100", ns, slug, id)
	err := c.http.GetAllJSON(path, func(body []byte) error {
		var page PagedResponse[wireServerPRActivity]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, a := range page.Values {
			if a.Action != "COMMENTED" || a.Comment.ID == 0 {
				continue
			}
			out = append(out, a.Comment.toDomain())
		}
		return nil
	})
	return out, err
}

type wireServerAddPRComment struct {
	Text string `json:"text"`
}

// AddPRComment adds a top-level comment to a pull request.
func (c *Client) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	body := wireServerAddPRComment{Text: in.Text}
	var w wireServerPRComment
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/comments", ns, slug, id)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PRComment{}, err
	}
	return w.toDomain(), nil
}
