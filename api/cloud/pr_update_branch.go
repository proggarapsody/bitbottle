package cloud

import "fmt"

// UpdatePRBranch syncs a PR's source branch with its base branch.
// POST /2.0/repositories/{workspace}/{slug}/pullrequests/{prID}/update-branch
// 202 Accepted on success.
func (c *Client) UpdatePRBranch(ns, slug string, prID int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/update-branch", ns, slug, prID)
	var result struct{}
	return c.postJSON(path, struct{}{}, &result)
}
