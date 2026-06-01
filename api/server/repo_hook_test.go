package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

// ── fixtures ──────────────────────────────────────────────────────────────────

const hooksPageJSON = `{
  "values": [
    {
      "details": {"key": "com.example.plugin:my-hook", "name": "My Hook", "version": "1.0.0"},
      "enabled": true,
      "configured": true
    },
    {
      "details": {"key": "com.example.plugin:other-hook", "name": "Other Hook", "version": "2.1.0"},
      "enabled": false,
      "configured": false
    }
  ],
  "size": 2,
  "isLastPage": true,
  "start": 0
}`

const singleHookJSON = `{
  "details": {"key": "com.example.plugin:my-hook", "name": "My Hook", "version": "1.0.0"},
  "enabled": true,
  "configured": true
}`

const hookSettingsJSON = `{"timeout": 30, "branch": "main"}`

// ── helpers ───────────────────────────────────────────────────────────────────

func newHookClient(t *testing.T, handler http.Handler) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "alice")
}

// ── ListRepoHooks ─────────────────────────────────────────────────────────────

func TestServerClient_ListRepoHooks(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(hooksPageJSON))
	}))

	got, err := client.ListRepoHooks("PROJ", "repo")
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks", seenPath)
	h := got[0]
	assert.Equal(t, "com.example.plugin:my-hook", h.Key)
	assert.Equal(t, "My Hook", h.Name)
	assert.Equal(t, "1.0.0", h.Version)
	assert.True(t, h.Enabled)
	assert.True(t, h.Configured)

	h2 := got[1]
	assert.Equal(t, "com.example.plugin:other-hook", h2.Key)
	assert.False(t, h2.Enabled)
	assert.False(t, h2.Configured)
}

// ── GetRepoHook ───────────────────────────────────────────────────────────────

func TestServerClient_GetRepoHook(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleHookJSON))
	}))

	got, err := client.GetRepoHook("PROJ", "repo", "com.example.plugin:my-hook")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks/com.example.plugin:my-hook", seenPath)
	assert.Equal(t, "com.example.plugin:my-hook", got.Key)
	assert.Equal(t, "My Hook", got.Name)
	assert.True(t, got.Enabled)
}

// ── EnableRepoHook ────────────────────────────────────────────────────────────

func TestServerClient_EnableRepoHook(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.WriteHeader(http.StatusOK)
	}))

	err := client.EnableRepoHook("PROJ", "repo", "com.example.plugin:my-hook")
	require.NoError(t, err)
	assert.Equal(t, "PUT", seenMethod)
	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks/com.example.plugin:my-hook/enabled", seenPath)
	assert.Equal(t, true, seenBody["enabled"])
}

// ── DisableRepoHook ───────────────────────────────────────────────────────────

func TestServerClient_DisableRepoHook(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.WriteHeader(http.StatusOK)
	}))

	err := client.DisableRepoHook("PROJ", "repo", "com.example.plugin:my-hook")
	require.NoError(t, err)
	assert.Equal(t, "PUT", seenMethod)
	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks/com.example.plugin:my-hook/enabled", seenPath)
	assert.Equal(t, false, seenBody["enabled"])
}

// ── GetRepoHookSettings ───────────────────────────────────────────────────────

func TestServerClient_GetRepoHookSettings(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(hookSettingsJSON))
	}))

	got, err := client.GetRepoHookSettings("PROJ", "repo", "com.example.plugin:my-hook")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks/com.example.plugin:my-hook/settings", seenPath)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(got, &parsed))
	assert.Equal(t, float64(30), parsed["timeout"])
	assert.Equal(t, "main", parsed["branch"])
}

// ── SetRepoHookSettings ───────────────────────────────────────────────────────

func TestServerClient_SetRepoHookSettings(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newHookClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(hookSettingsJSON))
	}))

	cfg := json.RawMessage(`{"timeout": 60, "branch": "main"}`)
	err := client.SetRepoHookSettings("PROJ", "repo", "com.example.plugin:my-hook", cfg)
	require.NoError(t, err)
	assert.Equal(t, "PUT", seenMethod)
	assert.Equal(t, "/projects/PROJ/repos/repo/settings/hooks/com.example.plugin:my-hook/settings", seenPath)
	assert.Equal(t, float64(60), seenBody["timeout"])
}
