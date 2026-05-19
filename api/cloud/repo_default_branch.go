package cloud

import "fmt"

// SetRepoDefaultBranch sets the default (main) branch of a repository.
// Cloud: PUT /2.0/repositories/{workspace}/{slug}
// body: {"mainbranch": {"name": branch}}
func (c *Client) SetRepoDefaultBranch(ns, slug, branch string) error {
	body := struct {
		MainBranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
	}{}
	body.MainBranch.Name = branch
	var ignore struct{}
	return c.putJSON(fmt.Sprintf("/repositories/%s/%s", ns, slug), body, &ignore)
}
