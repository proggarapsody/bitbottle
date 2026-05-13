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
