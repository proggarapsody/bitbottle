package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
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

// ListTags lists tags for a repository, following all pagination pages.
func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	var tags []backend.Tag
	path := fmt.Sprintf("/projects/%s/repos/%s/tags?limit=%d", ns, slug, limit)
	err := c.http.GetAllJSON(path, func(body []byte) error {
		var page PagedResponse[wireServerTag]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, w := range page.Values {
			tags = append(tags, w.toDomain())
		}
		return nil
	})
	return tags, err
}

type wireServerCreateTag struct {
	Name       string `json:"name"`
	StartPoint string `json:"startPoint"`
	Message    string `json:"message,omitempty"`
}

// CreateTag creates a new tag in a repository.
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

// DeleteTag deletes a tag in a repository.
func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/tags/%s", ns, slug, name)
	return c.delete(path, map[string]string{"name": name})
}
