package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// wireCloudCurrentUser is the minimal wire shape for GET /user.
type wireCloudCurrentUser struct {
	Nickname  string `json:"nickname"`
	AccountID string `json:"account_id"`
}

// currentUsername fetches and caches the current user's nickname via GET /user.
// The result is cached in c.cachedUsername so subsequent calls within the same
// command invocation avoid a redundant HTTP round-trip.
func (c *Client) currentUsername() (string, error) {
	c.cachedUsernameMu.Lock()
	defer c.cachedUsernameMu.Unlock()

	if c.cachedUsername != "" {
		return c.cachedUsername, nil
	}

	var u wireCloudCurrentUser
	if err := c.getJSON("/user", &u); err != nil {
		return "", fmt.Errorf("get current user: %w", err)
	}
	name := u.Nickname
	if name == "" {
		name = u.AccountID
	}
	if name == "" {
		return "", fmt.Errorf("could not determine current user identity from /user response")
	}
	c.cachedUsername = name
	return name, nil
}

// wireCloudSSHKey is the Cloud wire shape for a user SSH key.
type wireCloudSSHKey struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
}

func (w wireCloudSSHKey) toDomain() backend.SSHKey {
	return backend.SSHKey{
		ID:    w.ID,
		Label: w.Label,
		Key:   w.Key,
	}
}

// ListSSHKeys returns all SSH keys for the current user.
// GET /users/{username}/ssh-keys
func (c *Client) ListSSHKeys() ([]backend.SSHKey, error) {
	username, err := c.currentUsername()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/users/%s/ssh-keys", url.PathEscape(username))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.SSHKey, error) {
		var page cloudPagedResponse[wireCloudSSHKey]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.SSHKey, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, w.toDomain())
		}
		return out, nil
	}, 0)
}

// addSSHKeyBody is the request body for creating a user SSH key.
type addSSHKeyBody struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// AddSSHKey adds an SSH key for the current user.
// POST /users/{username}/ssh-keys
func (c *Client) AddSSHKey(input backend.SSHKeyInput) (backend.SSHKey, error) {
	if input.Key == "" {
		return backend.SSHKey{}, fmt.Errorf("key required")
	}
	username, err := c.currentUsername()
	if err != nil {
		return backend.SSHKey{}, err
	}
	body := addSSHKeyBody{Key: input.Key, Label: input.Label}
	var w wireCloudSSHKey
	path := fmt.Sprintf("/users/%s/ssh-keys", url.PathEscape(username))
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.SSHKey{}, err
	}
	return w.toDomain(), nil
}

// DeleteSSHKey removes an SSH key for the current user.
// DELETE /users/{username}/ssh-keys/{id}
func (c *Client) DeleteSSHKey(id int) error {
	username, err := c.currentUsername()
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/ssh-keys/%d", url.PathEscape(username), id)
	return c.delete(path)
}
