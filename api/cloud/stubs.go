package cloud

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

// UpdatePR stub — implemented by Scope G.
func (c *Client) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	return backend.PullRequest{}, fmt.Errorf("not implemented")
}

// DeclinePR stub — implemented by Scope G.
func (c *Client) DeclinePR(ns, slug string, id int) error {
	return fmt.Errorf("not implemented")
}

// UnapprovePR stub — implemented by Scope G.
func (c *Client) UnapprovePR(ns, slug string, id int) error {
	return fmt.Errorf("not implemented")
}

// ReadyPR stub — implemented by Scope G.
func (c *Client) ReadyPR(ns, slug string, id int) error {
	return fmt.Errorf("not implemented")
}

// RequestReview stub — implemented by Scope G.
func (c *Client) RequestReview(ns, slug string, id int, users []string) error {
	return fmt.Errorf("not implemented")
}
