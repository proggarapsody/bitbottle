package server

import "fmt"

// SetRepoDefaultBranch sets the default branch of a repository.
// Server: PUT /rest/api/1.0/projects/{project}/repos/{slug}
// body: {"defaultBranch": "branch"}
func (c *Client) SetRepoDefaultBranch(ns, slug, branch string) error {
	body := struct {
		DefaultBranch string `json:"defaultBranch"`
	}{DefaultBranch: branch}
	var ignore struct{}
	return c.putJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), body, &ignore)
}
