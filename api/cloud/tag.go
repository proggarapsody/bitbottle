package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
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

func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags?pagelen=%d", ns, slug, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Tag, error) {
		var page cloudPagedResponse[wireCloudTag]
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

type wireCloudCreateTag struct {
	Name    string             `json:"name"`
	Target  wireCloudTagTarget `json:"target"`
	Message string             `json:"message,omitempty"`
}

type wireCloudTagTarget struct {
	Hash string `json:"hash"`
}

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

func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags/%s", ns, slug, url.PathEscape(name))
	return c.delete(path)
}
