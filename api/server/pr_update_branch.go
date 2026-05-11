package server

import "fmt"

// UpdatePRBranch rebases a PR's source branch onto its base branch.
// POST /rest/api/1.0/projects/{key}/repos/{slug}/pull-requests/{prID}/rebase
// 200 OK on success.
func (c *Client) UpdatePRBranch(ns, slug string, prID int) error {
	path := fmt.Sprintf("/projects/%s/repos/%s/pull-requests/%d/rebase", ns, slug, prID)
	var result struct{}
	return c.postJSON(path, struct{}{}, &result)
}
