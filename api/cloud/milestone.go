package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

type cloudMilestone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func toMilestoneDomain(m cloudMilestone) backend.Milestone {
	return backend.Milestone{
		ID:   m.ID,
		Name: m.Name,
	}
}

// ListMilestones returns all milestones for a repository.
func (c *Client) ListMilestones(ns, slug string, limit int) ([]backend.Milestone, error) {
	path := fmt.Sprintf("/repositories/%s/%s/milestones",
		url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Milestone, error) {
		var page cloudPagedResponse[cloudMilestone]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Milestone, 0, len(page.Values))
		for _, m := range page.Values {
			out = append(out, toMilestoneDomain(m))
		}
		return out, nil
	}, limit)
}

// GetMilestone returns a single milestone by ID.
func (c *Client) GetMilestone(ns, slug string, id int) (backend.Milestone, error) {
	var m cloudMilestone
	path := fmt.Sprintf("/repositories/%s/%s/milestones/%d",
		url.PathEscape(ns), url.PathEscape(slug), id)
	if err := c.getJSON(path, &m); err != nil {
		return backend.Milestone{}, err
	}
	return toMilestoneDomain(m), nil
}
