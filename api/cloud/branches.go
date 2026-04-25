package cloud

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudBranch struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
}

func (w wireCloudBranch) toDomain() backend.Branch {
	return backend.Branch{
		Name:       w.Name,
		IsDefault:  false, // Cloud branch list doesn't include isDefault; repo.mainbranch would require a separate call
		LatestHash: w.Target.Hash,
	}
}

// ListBranches lists branches for a repository.
func (c *Client) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	var page cloudPagedResponse[wireCloudBranch]
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=%d", ns, slug, limit)
	if err := c.getJSON(path, &page); err != nil {
		return nil, err
	}
	branches := make([]backend.Branch, 0, len(page.Values))
	for _, w := range page.Values {
		branches = append(branches, w.toDomain())
	}
	return branches, nil
}
