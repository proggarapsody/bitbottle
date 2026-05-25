package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// restSSHKey is the wire representation of a Bitbucket Server/DC user SSH key.
type restSSHKey struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	// Text is the public key text on Server/DC (the field is named "text", not "key").
	Text string `json:"text"`
}

// restSSHKeyCreate is the request body for adding a user SSH key on Server/DC.
type restSSHKeyCreate struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

func toSSHKeyDomain(w restSSHKey) backend.SSHKey {
	return backend.SSHKey{
		ID:    w.ID,
		Label: w.Label,
		Key:   w.Text,
	}
}

// ListSSHKeys returns all SSH keys for the current user.
// GET /rest/ssh/1.0/keys
func (c *Client) ListSSHKeys() ([]backend.SSHKey, error) {
	return paging.Collect(c.sshHTTP, "/keys", func(body []byte) ([]backend.SSHKey, error) {
		var page PagedResponse[restSSHKey]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.SSHKey, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toSSHKeyDomain(w))
		}
		return out, nil
	}, 0)
}

// AddSSHKey adds an SSH key for the current user.
// POST /rest/ssh/1.0/keys
func (c *Client) AddSSHKey(input backend.SSHKeyInput) (backend.SSHKey, error) {
	if input.Key == "" {
		return backend.SSHKey{}, fmt.Errorf("key required")
	}
	body := restSSHKeyCreate{Text: input.Key, Label: input.Label}
	var w restSSHKey
	if err := c.sshHTTP.PostJSON("/keys", body, &w); err != nil {
		return backend.SSHKey{}, err
	}
	return toSSHKeyDomain(w), nil
}

// DeleteSSHKey removes an SSH key by ID for the current user.
// DELETE /rest/ssh/1.0/keys/{id}
func (c *Client) DeleteSSHKey(id int) error {
	return c.sshHTTP.DeleteJSON(fmt.Sprintf("/keys/%s", url.PathEscape(fmt.Sprintf("%d", id))), nil)
}
