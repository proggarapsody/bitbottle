package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	cloudgen "github.com/proggarapsody/bitbottle/api/cloud/gen"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// ListPRFiles returns the files changed in a pull request, paginated.
// Cloud endpoint: GET /repositories/{ws}/{slug}/pullrequests/{id}/diffstat
func (c *Client) ListPRFiles(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diffstat?pagelen=100", ns, slug, prID)
	return paging.Collect(c.http, path, func(body []byte) ([]backend.DiffStatEntry, error) {
		var page cloudPagedResponse[cloudgen.CloudDiffStatEntry]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.DiffStatEntry, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toDiffStatEntryDomain(w))
		}
		return out, nil
	}, 0)
}
