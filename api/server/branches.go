package server

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireBranch struct {
	ID           string `json:"id"`
	DisplayID    string `json:"displayId"`
	IsDefault    bool   `json:"isDefault"`
	LatestCommit string `json:"latestCommit"`
}

func (w wireBranch) toDomain() backend.Branch {
	return backend.Branch{
		Name:       w.DisplayID,
		IsDefault:  w.IsDefault,
		LatestHash: w.LatestCommit,
	}
}

// ListBranches lists branches for a repository, following all pagination pages.
func (c *Client) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	var branches []backend.Branch
	path := fmt.Sprintf("/projects/%s/repos/%s/branches?limit=%d", ns, slug, limit)
	err := c.http.GetAllJSON(path, func(body []byte) error {
		var page PagedResponse[wireBranch]
		if err := json.Unmarshal(body, &page); err != nil {
			return err
		}
		for _, w := range page.Values {
			branches = append(branches, w.toDomain())
		}
		return nil
	})
	return branches, err
}

// CreateBranch creates a new branch in the given repository.
// Server accepts both branch names and commit hashes in startPoint.
func (c *Client) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	type createRequest struct {
		Name       string `json:"name"`
		StartPoint string `json:"startPoint"`
	}
	req := createRequest{
		Name:       in.Name,
		StartPoint: in.StartAt,
	}
	path := fmt.Sprintf("/projects/%s/repos/%s/branches", ns, slug)
	var wire wireBranch
	if err := c.postJSON(path, req, &wire); err != nil {
		return backend.Branch{}, err
	}
	return wire.toDomain(), nil
}
