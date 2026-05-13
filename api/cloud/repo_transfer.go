package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

type wireCloudTransferBody struct {
	NewOwner wireCloudTransferOwner `json:"new_owner"`
}

type wireCloudTransferOwner struct {
	Username string `json:"username"`
}

// TransferRepo transfers a Cloud repository to another workspace.
// POST /repositories/{workspace}/{slug}/transfer
func (c *Client) TransferRepo(ns, slug, target string) (backend.Repository, error) {
	body := wireCloudTransferBody{
		NewOwner: wireCloudTransferOwner{Username: target},
	}
	var w wireCloudRepo
	path := fmt.Sprintf("/repositories/%s/%s/transfer", url.PathEscape(ns), url.PathEscape(slug))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Repository{}, err
	}
	return w.toDomain(), nil
}
