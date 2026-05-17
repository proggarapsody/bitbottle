package server

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// TransferRepo transfers a Server repository to another project.
// PUT /rest/api/1.0/projects/{ns}/repos/{slug}
func (c *Client) TransferRepo(ns, slug, target string) (backend.Repository, error) {
	body := servergen.RestTransferBody{
		Project: servergen.RestTransferProject{Key: target},
	}
	var w servergen.RestRepository
	if err := c.putJSON(fmt.Sprintf("/projects/%s/repos/%s", ns, slug), body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepositoryDomain(w), nil
}
