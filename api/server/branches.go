package server

import (
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

// ListBranches lists branches for a repository.
func (c *Client) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	var page PagedResponse[wireBranch]
	path := fmt.Sprintf("/projects/%s/repos/%s/branches?limit=%d", ns, slug, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	branches := make([]backend.Branch, 0, len(page.Values))
	for _, w := range page.Values {
		branches = append(branches, w.toDomain())
	}
	return branches, nil
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
