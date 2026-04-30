package cloud

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// hexHashRE matches a 40-character lowercase hexadecimal commit hash.
var hexHashRE = regexp.MustCompile(`^[0-9a-f]{40}$`)

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

// ListBranches lists branches for a repository, capping the total returned
// at limit (limit <= 0 means unbounded).
func (c *Client) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=%d", ns, slug, limit)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.Branch, error) {
		var page cloudPagedResponse[wireCloudBranch]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.Branch, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, limit)
}

// CreateBranch creates a new branch in the given repository.
// Cloud requires a commit hash as the target; if StartAt is a branch name
// (not a 40-char hex string) it is resolved to its HEAD commit hash first.
func (c *Client) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	hash := in.StartAt
	if !hexHashRE.MatchString(hash) {
		// StartAt looks like a branch name — resolve it to a commit hash.
		resolvePath := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", ns, slug, hash)
		var resolved wireCloudBranch
		if err := c.getJSON(resolvePath, &resolved); err != nil {
			return backend.Branch{}, fmt.Errorf("resolve branch %q: %w", hash, err)
		}
		hash = resolved.Target.Hash
	}

	type createRequest struct {
		Name   string `json:"name"`
		Target struct {
			Hash string `json:"hash"`
		} `json:"target"`
	}

	var req createRequest
	req.Name = in.Name
	req.Target.Hash = hash

	path := fmt.Sprintf("/repositories/%s/%s/refs/branches", ns, slug)
	var wire wireCloudBranch
	if err := c.postJSON(path, req, &wire); err != nil {
		return backend.Branch{}, err
	}
	return wire.toDomain(), nil
}
