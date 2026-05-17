package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

func toTagDomain(w cloudgen.CloudTag) backend.Tag {
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
		var page cloudPagedResponse[cloudgen.CloudTag]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Tag, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toTagDomain(w))
		}
		return out, nil
	}, limit)
}

func (c *Client) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	body := cloudgen.CloudCreateTag{
		Name:   in.Name,
		Target: cloudgen.CloudTagTarget{Hash: in.StartAt},
	}
	if in.Message != "" {
		body.Message = &in.Message
	}
	var w cloudgen.CloudTag
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Tag{}, err
	}
	return toTagDomain(w), nil
}

func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/tags/%s", ns, slug, url.PathEscape(name))
	return c.delete(path)
}
