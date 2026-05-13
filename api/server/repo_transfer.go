package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireServerTransferBody struct {
	Project wireServerTransferProject `json:"project"`
}

type wireServerTransferProject struct {
	Key string `json:"key"`
}

// TransferRepo transfers a Server repository to another project.
// PUT /rest/api/1.0/projects/{ns}/repos/{slug}
func (c *Client) TransferRepo(ns, slug, target string) (backend.Repository, error) {
	body := wireServerTransferBody{
		Project: wireServerTransferProject{Key: target},
	}
	var w wireRepository
	if err := c.putJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}
