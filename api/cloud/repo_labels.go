package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// cloudRepoLabel is the wire shape for a Bitbucket Cloud repository label.
type cloudRepoLabel struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func toRepoLabelDomain(w cloudRepoLabel) backend.RepoLabel {
	return backend.RepoLabel{
		ID:    w.ID,
		Name:  w.Name,
		Color: w.Color,
	}
}

// ListRepoLabels fetches all labels for a Cloud repository.
func (c *Client) ListRepoLabels(ns, slug string) ([]backend.RepoLabel, error) {
	path := fmt.Sprintf("/repositories/%s/%s/labels",
		url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.RepoLabel, error) {
		var page cloudPagedResponse[cloudRepoLabel]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.RepoLabel, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toRepoLabelDomain(w))
		}
		return out, nil
	}, 0)
}

// CreateRepoLabel creates a new label on a Cloud repository.
func (c *Client) CreateRepoLabel(ns, slug string, in backend.CreateRepoLabelInput) (backend.RepoLabel, error) {
	path := fmt.Sprintf("/repositories/%s/%s/labels",
		url.PathEscape(ns), url.PathEscape(slug))
	body := cloudRepoLabel{Name: in.Name, Color: in.Color}
	var w cloudRepoLabel
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.RepoLabel{}, err
	}
	return toRepoLabelDomain(w), nil
}

// UpdateRepoLabel updates an existing label on a Cloud repository.
func (c *Client) UpdateRepoLabel(ns, slug string, id int, in backend.UpdateRepoLabelInput) (backend.RepoLabel, error) {
	path := fmt.Sprintf("/repositories/%s/%s/labels/%d",
		url.PathEscape(ns), url.PathEscape(slug), id)
	body := cloudRepoLabel{Name: in.Name, Color: in.Color}
	var w cloudRepoLabel
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.RepoLabel{}, err
	}
	return toRepoLabelDomain(w), nil
}

// DeleteRepoLabel removes a label from a Cloud repository.
func (c *Client) DeleteRepoLabel(ns, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/labels/%d",
		url.PathEscape(ns), url.PathEscape(slug), id)
	return c.delete(path)
}
