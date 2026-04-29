package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudTag struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
	Message string `json:"message"`
	Links   struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
}

func (w wireCloudTag) toDomain() backend.Tag {
	return backend.Tag{
		Name:    w.Name,
		Hash:    w.Target.Hash,
		Message: w.Message,
		WebURL:  w.Links.HTML.Href,
	}
}

// ListTags lists tags for a repository, following all pagination pages.
func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	var tags []backend.Tag
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags?pagelen=%d", ns, slug, limit)
	err := c.http.GetAllJSON(path, func(body []byte) error {
		var page cloudPagedResponse[wireCloudTag]
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

type wireCloudCreateTag struct {
	Name    string             `json:"name"`
	Target  wireCloudTagTarget `json:"target"`
	Message string             `json:"message,omitempty"`
}

type wireCloudTagTarget struct {
	Hash string `json:"hash"`
}

// CreateTag creates a new tag in a repository.
func (c *Client) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	body := wireCloudCreateTag{
		Name:    in.Name,
		Target:  wireCloudTagTarget{Hash: in.StartAt},
		Message: in.Message,
	}
	var w wireCloudTag
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Tag{}, err
	}
	return w.toDomain(), nil
}

// DeleteTag deletes a tag in a repository.
func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags/%s", ns, slug, url.PathEscape(name))
	return c.delete(path)
}
