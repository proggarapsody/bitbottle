package server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
	servergen "github.com/proggarapsody/bitbottle/api/server/gen"
)

// ── AdminClient implementation ────────────────────────────────────────────────

// RotateSecrets rotates the cluster's internal HTTPS secret.
// POST /rest/api/1.0/admin/secrets — no request body; 200 OK = success.
func (c *Client) RotateSecrets() error {
	return c.http.PostJSON("/admin/secrets", nil, nil)
}

// GetLoggingConfig returns the current log level and async-logging setting.
// GET /rest/api/1.0/admin/logging
func (c *Client) GetLoggingConfig() (backend.LoggingConfig, error) {
	var wire servergen.RestLoggingConfig
	if err := c.http.GetJSON("/admin/logging", &wire); err != nil {
		return backend.LoggingConfig{}, err
	}
	return backend.LoggingConfig{
		Level: wire.LogLevel,
		Async: wire.AsyncLogging,
	}, nil
}

// SetLoggingConfig updates the log level and/or async-logging setting.
// If in.Persistent is true: PUT /rest/api/1.0/admin/logging/properties
// Otherwise:                PUT /rest/api/1.0/admin/logging
func (c *Client) SetLoggingConfig(in backend.LoggingConfigInput) error {
	wire := servergen.RestLoggingConfig{
		LogLevel:     in.Level,
		AsyncLogging: in.Async,
	}
	path := "/admin/logging"
	if in.Persistent {
		path = "/admin/logging/properties"
	}
	return c.http.PutJSON(path, wire, nil)
}

// ── User management ───────────────────────────────────────────────────────────

// adminUserWire is the wire representation of a user from the admin users
// endpoint.
type adminUserWire struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
	Type         string `json:"type"`
}

// ListAdminUsers returns admin users on Bitbucket Server/DC, optionally
// filtered by a query string.
// GET /rest/api/1.0/admin/users?filter=QUERY&limit=N&start=P
func (c *Client) ListAdminUsers(filter string, limit int) ([]backend.AdminUser, error) {
	path := fmt.Sprintf("/admin/users?limit=%d", limit)
	if filter != "" {
		path += "&filter=" + url.QueryEscape(filter)
	}
	return paging.Collect(c.http, path, func(body []byte) ([]backend.AdminUser, error) {
		var page PagedResponse[adminUserWire]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.AdminUser, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, backend.AdminUser{
				Slug:        w.Name,
				DisplayName: w.DisplayName,
				Email:       w.EmailAddress,
				Active:      w.Active,
				Type:        w.Type,
			})
		}
		return out, nil
	}, limit)
}

// RenameUser renames a user (changes their username/slug).
// PUT /rest/api/1.0/admin/users with body {"name": "new-slug"}
func (c *Client) RenameUser(slug, newSlug string) error {
	type renameRequest struct {
		Name    string `json:"name"`
		NewName string `json:"newName"`
	}
	return c.http.PutJSON("/admin/users/rename", renameRequest{Name: slug, NewName: newSlug}, nil)
}

// ActivateUser activates a user account.
// PUT /rest/api/1.0/admin/users with body {"active": true}
func (c *Client) ActivateUser(slug string) error {
	type activateRequest struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	return c.http.PutJSON("/admin/users", activateRequest{Name: slug, Active: true}, nil)
}

// DeactivateUser deactivates a user account.
// PUT /rest/api/1.0/admin/users with body {"active": false}
func (c *Client) DeactivateUser(slug string) error {
	type deactivateRequest struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	return c.http.PutJSON("/admin/users", deactivateRequest{Name: slug, Active: false}, nil)
}

// ── System info ───────────────────────────────────────────────────────────────

// GetLicense returns the current license details for the instance.
// GET /rest/api/1.0/admin/license
func (c *Client) GetLicense() (backend.AdminLicense, error) {
	var wire struct {
		Tier              string `json:"tier"`
		Users             int    `json:"numberOfUsers"`
		ServerId          string `json:"serverId"`
		License           string `json:"license"`
		PurchaseDate      string `json:"purchaseDate"`
		ExpiryDate        string `json:"expiryDate"`
		SupportExpiryDate string `json:"supportExpiryDate"`
		CreationDate      string `json:"creationDate"`
	}
	if err := c.http.GetJSON("/admin/license", &wire); err != nil {
		return backend.AdminLicense{}, err
	}
	return backend.AdminLicense{
		Tier:          wire.Tier,
		Users:         wire.Users,
		ServerId:      wire.ServerId,
		License:       wire.License,
		ExpiryDate:    wire.ExpiryDate,
		SupportExpiry: wire.SupportExpiryDate,
		CreationDate:  wire.CreationDate,
	}, nil
}

// GetMailServerConfig returns the current mail-server configuration.
// GET /rest/api/1.0/admin/mail-server
func (c *Client) GetMailServerConfig() (backend.MailServerConfig, error) {
	var wire struct {
		Hostname        string `json:"hostname"`
		Port            int    `json:"port"`
		Protocol        string `json:"protocol"`
		UseStartTLS     bool   `json:"use-start-tls"`
		RequireStartTLS bool   `json:"require-start-tls"`
		Username        string `json:"username"`
		SenderAddress   string `json:"senderAddress"`
	}
	if err := c.http.GetJSON("/admin/mail-server", &wire); err != nil {
		return backend.MailServerConfig{}, err
	}
	return backend.MailServerConfig{
		Hostname:        wire.Hostname,
		Port:            wire.Port,
		Protocol:        wire.Protocol,
		UseStartTLS:     wire.UseStartTLS,
		RequireStartTLS: wire.RequireStartTLS,
		Username:        wire.Username,
		SenderAddress:   wire.SenderAddress,
	}, nil
}

// SetMailServerConfig writes a new mail-server configuration.
// PUT /rest/api/1.0/admin/mail-server
func (c *Client) SetMailServerConfig(in backend.MailServerConfig) error {
	wire := struct {
		Hostname        string `json:"hostname"`
		Port            int    `json:"port"`
		Protocol        string `json:"protocol"`
		UseStartTLS     bool   `json:"use-start-tls"`
		RequireStartTLS bool   `json:"require-start-tls"`
		Username        string `json:"username"`
		SenderAddress   string `json:"senderAddress"`
		Password        string `json:"password,omitempty"`
	}{
		Hostname:        in.Hostname,
		Port:            in.Port,
		Protocol:        in.Protocol,
		UseStartTLS:     in.UseStartTLS,
		RequireStartTLS: in.RequireStartTLS,
		Username:        in.Username,
		SenderAddress:   in.SenderAddress,
		Password:        in.Password,
	}
	return c.http.PutJSON("/admin/mail-server", wire, nil)
}

// GetClusterNodes returns the nodes in the Bitbucket Server/DC cluster.
// GET /rest/api/1.0/admin/cluster
func (c *Client) GetClusterNodes() ([]backend.ClusterNode, error) {
	var wire struct {
		Nodes []struct {
			NodeId  string `json:"nodeId"`
			Name    string `json:"name"`
			Address struct {
				Address string `json:"address"`
			} `json:"address"`
			State string `json:"state"`
			Local bool   `json:"local"`
		} `json:"nodes"`
	}
	if err := c.http.GetJSON("/admin/cluster", &wire); err != nil {
		return nil, err
	}
	out := make([]backend.ClusterNode, 0, len(wire.Nodes))
	for _, n := range wire.Nodes {
		out = append(out, backend.ClusterNode{
			NodeId:  n.NodeId,
			Name:    n.Name,
			Address: n.Address.Address,
			State:   n.State,
			Local:   n.Local,
		})
	}
	return out, nil
}
