package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// CreateBranch stub — implemented by Scope L.
func (c *Client) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	return backend.Branch{}, fmt.Errorf("not implemented")
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
