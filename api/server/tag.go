package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type wireServerTag struct {
	DisplayID      string `json:"displayId"`
	LatestCommit   string `json:"latestCommit"`
	DisplayMessage string `json:"displayMessage"`
}

func (w wireServerTag) toDomain() backend.Tag {
	return backend.Tag{
		Name:    w.DisplayID,
		Hash:    w.LatestCommit,
		Message: w.DisplayMessage,
	}
}

func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/tags?limit=%d", ns, slug, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Tag, error) {
		var page PagedResponse[wireServerTag]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Tag, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

type wireServerCreateTag struct {
	Name       string `json:"name"`
	StartPoint string `json:"startPoint"`
	Message    string `json:"message,omitempty"`
}

func (c *Client) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	body := wireServerCreateTag{
		Name:       in.Name,
		StartPoint: in.StartAt,
		Message:    in.Message,
	}
	var w wireServerTag
	path := fmt.Sprintf("/projects/%s/repos/%s/tags", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Tag{}, err
	}
	return w.toDomain(), nil
}

func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/tags/%s", ns, slug, name)
	return c.delete(path, map[string]string{"name": name})
}
