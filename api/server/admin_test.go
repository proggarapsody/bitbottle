package server_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── RotateSecrets ─────────────────────────────────────────────────────────────

func TestServerClient_RotateSecrets_OK(t *testing.T) {
	t.Parallel()
	var called bool
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/admin/secrets", r.URL.Path)
		called = true
		w.WriteHeader(http.StatusOK)
	})
	err := c.RotateSecrets()
	require.NoError(t, err)
	assert.True(t, called)
}

func TestServerClient_RotateSecrets_403_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := c.RotateSecrets()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrPermission, de.Kind)
}

// ── GetLoggingConfig ──────────────────────────────────────────────────────────

func TestServerClient_GetLoggingConfig_OK(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/admin/logging", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"logLevel":"DEBUG","asyncLogging":true}`))
	})
	cfg, err := c.GetLoggingConfig()
	require.NoError(t, err)
	assert.Equal(t, "DEBUG", cfg.Level)
	assert.True(t, cfg.Async)
}

// ── SetLoggingConfig ──────────────────────────────────────────────────────────

func TestServerClient_SetLoggingConfig_RuntimeOnly(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.SetLoggingConfig(backend.LoggingConfigInput{Level: "WARN", Async: false, Persistent: false})
	require.NoError(t, err)
	assert.Equal(t, "/admin/logging", gotPath)
	assert.Contains(t, gotBody, `"logLevel":"WARN"`)
}

func TestServerClient_SetLoggingConfig_Persistent(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	err := c.SetLoggingConfig(backend.LoggingConfigInput{Level: "ERROR", Async: true, Persistent: true})
	require.NoError(t, err)
	assert.Equal(t, "/admin/logging/properties", gotPath)
}

// ── ListAdminUsers ────────────────────────────────────────────────────────────

func TestServerClient_ListAdminUsers_OK(t *testing.T) {
	t.Parallel()
	const responseJSON = `{
		"values": [
			{"name":"alice","displayName":"Alice A","emailAddress":"alice@example.com","active":true,"type":"NORMAL"},
			{"name":"svc-bot","displayName":"Bot","emailAddress":"bot@example.com","active":false,"type":"SERVICE"}
		],
		"isLastPage": true
	}`
	var gotPath string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})
	users, err := c.ListAdminUsers("", 50)
	require.NoError(t, err)
	assert.Equal(t, "/admin/users", gotPath)
	require.Len(t, users, 2)
	assert.Equal(t, "alice", users[0].Slug)
	assert.Equal(t, "Alice A", users[0].DisplayName)
	assert.Equal(t, "alice@example.com", users[0].Email)
	assert.True(t, users[0].Active)
	assert.Equal(t, "NORMAL", users[0].Type)
	assert.Equal(t, "svc-bot", users[1].Slug)
	assert.False(t, users[1].Active)
}

func TestServerClient_ListAdminUsers_WithFilter(t *testing.T) {
	t.Parallel()
	var gotQuery string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	})
	_, err := c.ListAdminUsers("ali", 25)
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "filter=ali")
	assert.Contains(t, gotQuery, "limit=25")
}

// ── ActivateUser / DeactivateUser ─────────────────────────────────────────────

func TestServerClient_ActivateUser_OK(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.ActivateUser("alice")
	require.NoError(t, err)
	assert.Equal(t, "/admin/users", gotPath)
	assert.Contains(t, gotBody, `"name":"alice"`)
	assert.Contains(t, gotBody, `"active":true`)
}

func TestServerClient_DeactivateUser_OK(t *testing.T) {
	t.Parallel()
	var gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.DeactivateUser("alice")
	require.NoError(t, err)
	assert.Contains(t, gotBody, `"name":"alice"`)
	assert.Contains(t, gotBody, `"active":false`)
}

// ── RenameUser ────────────────────────────────────────────────────────────────

func TestServerClient_RenameUser_OK(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.RenameUser("old-alice", "alice")
	require.NoError(t, err)
	assert.Equal(t, "/admin/users/rename", gotPath)
	assert.Contains(t, gotBody, `"name":"old-alice"`)
	assert.Contains(t, gotBody, `"newName":"alice"`)
}

// ── GetLicense ────────────────────────────────────────────────────────────────

func TestServerClient_GetLicense_OK(t *testing.T) {
	t.Parallel()
	const responseJSON = `{
		"tier": "ENTERPRISE",
		"numberOfUsers": 500,
		"serverId": "srv-abc123",
		"license": "ABC-LICENSE-KEY",
		"expiryDate": "2027-01-01",
		"supportExpiryDate": "2026-01-01",
		"creationDate": "2020-01-01"
	}`
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/admin/license", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})
	lic, err := c.GetLicense()
	require.NoError(t, err)
	assert.Equal(t, "ENTERPRISE", lic.Tier)
	assert.Equal(t, 500, lic.Users)
	assert.Equal(t, "srv-abc123", lic.ServerId)
	assert.Equal(t, "2027-01-01", lic.ExpiryDate)
	assert.Equal(t, "2026-01-01", lic.SupportExpiry)
}

// ── GetMailServerConfig ───────────────────────────────────────────────────────

func TestServerClient_GetMailServerConfig_OK(t *testing.T) {
	t.Parallel()
	const responseJSON = `{
		"hostname": "smtp.example.com",
		"port": 465,
		"protocol": "smtps",
		"use-start-tls": false,
		"require-start-tls": false,
		"username": "mailer",
		"senderAddress": "no-reply@example.com"
	}`
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/admin/mail-server", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})
	cfg, err := c.GetMailServerConfig()
	require.NoError(t, err)
	assert.Equal(t, "smtp.example.com", cfg.Hostname)
	assert.Equal(t, 465, cfg.Port)
	assert.Equal(t, "smtps", cfg.Protocol)
	assert.False(t, cfg.UseStartTLS)
	assert.False(t, cfg.RequireStartTLS)
	assert.Equal(t, "mailer", cfg.Username)
	assert.Equal(t, "no-reply@example.com", cfg.SenderAddress)
	// Password is never returned by GET — field must be zero value.
	assert.Equal(t, "", cfg.Password)
}

// ── SetMailServerConfig ───────────────────────────────────────────────────────

func TestServerClient_SetMailServerConfig_OK(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.SetMailServerConfig(backend.MailServerConfig{
		Hostname:      "smtp.example.com",
		Port:          25,
		Protocol:      "smtp",
		UseStartTLS:   true,
		Username:      "mailer",
		SenderAddress: "no-reply@example.com",
		Password:      "s3cr3t",
	})
	require.NoError(t, err)
	assert.Equal(t, "/admin/mail-server", gotPath)
	assert.Contains(t, gotBody, `"hostname":"smtp.example.com"`)
	assert.Contains(t, gotBody, `"port":25`)
	assert.Contains(t, gotBody, `"protocol":"smtp"`)
	assert.Contains(t, gotBody, `"use-start-tls":true`)
	assert.Contains(t, gotBody, `"password":"s3cr3t"`)
}

func TestServerClient_SetMailServerConfig_403_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := c.SetMailServerConfig(backend.MailServerConfig{
		Hostname: "smtp.example.com",
	})
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrPermission, de.Kind)
}

// ── GetBanner ─────────────────────────────────────────────────────────────────

func TestServerClient_GetBanner_OK(t *testing.T) {
	t.Parallel()
	const responseJSON = `{"message":"Maintenance on Friday","audience":"ALL","enabled":true}`
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/admin/banner", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})
	cfg, err := c.GetBanner()
	require.NoError(t, err)
	assert.Equal(t, "Maintenance on Friday", cfg.Message)
	assert.Equal(t, "ALL", cfg.Audience)
	assert.True(t, cfg.Enabled)
}

func TestServerClient_GetBanner_403_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	_, err := c.GetBanner()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrPermission, de.Kind)
}

// ── SetBanner ─────────────────────────────────────────────────────────────────

func TestServerClient_SetBanner_OK(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	err := c.SetBanner(backend.BannerConfig{
		Message:  "Scheduled downtime Sunday 02:00 UTC",
		Audience: "ALL",
		Enabled:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, "/admin/banner", gotPath)
	assert.Contains(t, gotBody, `"message":"Scheduled downtime Sunday 02:00 UTC"`)
	assert.Contains(t, gotBody, `"audience":"ALL"`)
	assert.Contains(t, gotBody, `"enabled":true`)
}

func TestServerClient_SetBanner_403_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := c.SetBanner(backend.BannerConfig{Message: "test", Audience: "ALL", Enabled: true})
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrPermission, de.Kind)
}

// ── ClearBanner ───────────────────────────────────────────────────────────────

func TestServerClient_ClearBanner_OK(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.ClearBanner()
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/admin/banner", gotPath)
}

func TestServerClient_ClearBanner_403_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := c.ClearBanner()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrPermission, de.Kind)
}

// ── GetClusterNodes ───────────────────────────────────────────────────────────

func TestServerClient_GetClusterNodes_OK(t *testing.T) {
	t.Parallel()
	const responseJSON = `{
		"nodes": [
			{"nodeId":"node-1","name":"Primary","address":{"address":"10.0.0.1"},"state":"ACTIVE","local":true},
			{"nodeId":"node-2","name":"Secondary","address":{"address":"10.0.0.2"},"state":"ACTIVE","local":false}
		]
	}`
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/admin/cluster", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	})
	nodes, err := c.GetClusterNodes()
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	assert.Equal(t, "node-1", nodes[0].NodeId)
	assert.Equal(t, "Primary", nodes[0].Name)
	assert.Equal(t, "10.0.0.1", nodes[0].Address)
	assert.Equal(t, "ACTIVE", nodes[0].State)
	assert.True(t, nodes[0].Local)
	assert.False(t, nodes[1].Local)
}
