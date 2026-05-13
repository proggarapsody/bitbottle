package server

import (
	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── wire types ───────────────────────────────────────────────────────────────

type wireLoggingConfig struct {
	LogLevel     string `json:"logLevel"`
	AsyncLogging bool   `json:"asyncLogging"`
}

// ── AdminClient implementation ────────────────────────────────────────────────

// RotateSecrets rotates the cluster's internal HTTPS secret.
// POST /rest/api/1.0/admin/secrets — no request body; 200 OK = success.
func (c *Client) RotateSecrets() error {
	return c.http.PostJSON("/admin/secrets", nil, nil)
}

// GetLoggingConfig returns the current log level and async-logging setting.
// GET /rest/api/1.0/admin/logging
func (c *Client) GetLoggingConfig() (backend.LoggingConfig, error) {
	var wire wireLoggingConfig
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
	wire := wireLoggingConfig{
		LogLevel:     in.Level,
		AsyncLogging: in.Async,
	}
	path := "/admin/logging"
	if in.Persistent {
		path = "/admin/logging/properties"
	}
	return c.http.PutJSON(path, wire, nil)
}
