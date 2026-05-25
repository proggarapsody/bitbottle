package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// serverRepoLabel is the wire shape for a Bitbucket Server/DC repository label.
type serverRepoLabel struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func toServerRepoLabelDomain(w serverRepoLabel) backend.RepoLabel {
	return backend.RepoLabel{
		ID:    w.ID,
		Name:  w.Name,
		Color: w.Color,
	}
}

// ListRepoLabels fetches all labels for a Server/DC repository.
func (c *Client) ListRepoLabels(ns, slug string) ([]backend.RepoLabel, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/labels", ns, slug)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.RepoLabel, error) {
		var page PagedResponse[serverRepoLabel]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.RepoLabel, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toServerRepoLabelDomain(w))
		}
		return out, nil
	}, 0)
}

// CreateRepoLabel creates a new label on a Server/DC repository.
func (c *Client) CreateRepoLabel(ns, slug string, in backend.CreateRepoLabelInput) (backend.RepoLabel, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/labels", ns, slug)
	body := serverRepoLabel{Name: in.Name, Color: in.Color}
	var w serverRepoLabel
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.RepoLabel{}, err
	}
	return toServerRepoLabelDomain(w), nil
}

// UpdateRepoLabel updates an existing label on a Server/DC repository.
func (c *Client) UpdateRepoLabel(ns, slug string, id int, in backend.UpdateRepoLabelInput) (backend.RepoLabel, error) {
	path := fmt.Sprintf("/projects/%s/repos/%s/labels/%d", ns, slug, id)
	body := serverRepoLabel{Name: in.Name, Color: in.Color}
	var w serverRepoLabel
	if err := c.putJSON(path, body, &w); err != nil {
		return backend.RepoLabel{}, err
	}
	return toServerRepoLabelDomain(w), nil
}

// DeleteRepoLabel removes a label from a Server/DC repository.
func (c *Client) DeleteRepoLabel(ns, slug string, id int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/labels/%d", ns, slug, id)
	return c.delete(path, nil)
}
