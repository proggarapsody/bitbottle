package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// CreateBranch stub — implemented by Scope L.
func (c *Client) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	return backend.Branch{}, fmt.Errorf("not implemented")
}

// ListTags stub — implemented by Scope E.
func (c *Client) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateTag stub — implemented by Scope E.
func (c *Client) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	return backend.Tag{}, fmt.Errorf("not implemented")
}

// DeleteTag stub — implemented by Scope E.
func (c *Client) DeleteTag(ns, slug, name string) error {
	return fmt.Errorf("not implemented")
}
