package server

import (
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

// ListTags lists tags for a repository.
func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	var page PagedResponse[wireServerTag]
	path := fmt.Sprintf("/projects/%s/repos/%s/tags?limit=%d", ns, slug, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	tags := make([]backend.Tag, 0, len(page.Values))
	for _, w := range page.Values {
		tags = append(tags, w.toDomain())
	}
	return tags, nil
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
