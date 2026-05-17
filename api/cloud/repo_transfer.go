package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
)

// TransferRepo transfers a Cloud repository to another workspace.
// POST /repositories/{workspace}/{slug}/transfer
func (c *Client) TransferRepo(ns, slug, target string) (backend.Repository, error) {
	body := cloudgen.CloudTransferBody{
		NewOwner: cloudgen.CloudTransferOwner{Username: target},
	}
	var w cloudgen.CloudRepo
	path := fmt.Sprintf("/repositories/%s/%s/transfer", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Repository{}, err
	}
	return toRepoDomain(w), nil
}
