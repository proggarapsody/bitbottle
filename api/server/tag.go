package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

func toTagDomain(w servergen.RestTag) backend.Tag {
	return backend.Tag{
		Name:    w.DisplayID,
		Hash:    w.LatestCommit,
		Message: w.DisplayMessage,
	}
}

func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/tags?limit=%d", ns, slug, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Tag, error) {
		var page PagedResponse[servergen.RestTag]
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
	body := servergen.RestCreateTag{
		Name:       in.Name,
		StartPoint: in.StartAt,
		Message:    in.Message,
	}
	var w servergen.RestTag
	path := fmt.Sprintf("/projects/%s/repos/%s/tags", ns, slug)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Tag{}, err
	}
	return toTagDomain(w), nil
}

func (c *Client) DeleteTag(ns, slug, name string) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/tags/%s", ns, slug, name)
	return c.delete(path, map[string]string{"name": name})
}
