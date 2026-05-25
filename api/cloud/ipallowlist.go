package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireIPAllowlist is the Bitbucket Cloud JSON shape for one IP allowlist entry.
type wireIPAllowlist struct {
	UUID        string `json:"uuid"`
	CIDR        string `json:"cidr"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

func toIPAllowlistDomain(w wireIPAllowlist) backend.IPAllowlist {
	return backend.IPAllowlist{
		UUID:        w.UUID,
		CIDR:        w.CIDR,
		Description: w.Description,
		Enabled:     w.Enabled,
	}
}

// ListIPAllowlists returns all IP allowlist entries for the given workspace.
func (c *Client) ListIPAllowlists(workspace string) ([]backend.IPAllowlist, error) {
	path := fmt.Sprintf("/workspaces/%s/ipallowlists", url.PathEscape(workspace))
	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.IPAllowlist, error) {
			var page struct {
				Values []wireIPAllowlist `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.IPAllowlist, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, toIPAllowlistDomain(w))
			}
			return out, nil
		},
		0, // unbounded
	)
}

// wireIPAllowlistCreate is the request body for POST /workspaces/{ws}/ipallowlists.
type wireIPAllowlistCreate struct {
	CIDR        string `json:"cidr"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// CreateIPAllowlist adds a new IP allowlist entry to a workspace.
func (c *Client) CreateIPAllowlist(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error) {
	body := wireIPAllowlistCreate{
		CIDR:        in.CIDR,
		Description: in.Description,
		Enabled:     in.Enabled,
	}
	var w wireIPAllowlist
	path := fmt.Sprintf("/workspaces/%s/ipallowlists", url.PathEscape(workspace))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.IPAllowlist{}, err
	}
	return toIPAllowlistDomain(w), nil
}

// DeleteIPAllowlist removes an IP allowlist entry from a workspace.
func (c *Client) DeleteIPAllowlist(workspace, entryUUID string) error {
	path := fmt.Sprintf("/workspaces/%s/ipallowlists/%s", url.PathEscape(workspace), url.PathEscape(entryUUID))
	return c.delete(path)
}
