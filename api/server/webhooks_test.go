package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func TestServerClient_ListWebhooks_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWebhooks("PROJ", "repo")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/repo/webhooks", gotPath)
}

func TestServerClient_ListWebhooks_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
            "values":[
                {"id":1,"name":"hook","url":"https://example.com/h1","active":true,"events":["repo:refs_changed","pr:opened"]},
                {"id":2,"name":"hook2","url":"https://example.com/h2","active":false,"events":["repo:refs_changed"]}
            ],
            "isLastPage":true
        }`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	hooks, err := client.ListWebhooks("PROJ", "repo")
	require.NoError(t, err)
	require.Len(t, hooks, 2)
	assert.Equal(t, "1", hooks[0].ID)
	assert.Equal(t, "https://example.com/h1", hooks[0].URL)
	assert.True(t, hooks[0].Active)
	assert.Equal(t, []string{"repo:refs_changed", "pr:opened"}, hooks[0].Events)
	assert.Equal(t, "2", hooks[1].ID)
	assert.False(t, hooks[1].Active)
}

func TestServerClient_GetWebhook_PathContainsID(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"name":"hook","url":"https://example.com/x","active":true,"events":["repo:refs_changed"]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	hook, err := client.GetWebhook("PROJ", "repo", "42")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/repo/webhooks/42", gotPath)
	assert.Equal(t, "42", hook.ID)
}

func TestServerClient_CreateWebhook_PostsExpectedBody(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"name":"bitbottle","url":"https://example.com/h","active":true,"events":["repo:refs_changed"]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	hook, err := client.CreateWebhook("PROJ", "repo", backend.CreateWebhookInput{
		URL:    "https://example.com/h",
		Events: []string{"repo:refs_changed"},
		Active: true,
		Secret: "redacted-test-value",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/projects/PROJ/repos/repo/webhooks", gotPath)
	assert.Equal(t, "https://example.com/h", gotBody["url"])
	assert.Equal(t, true, gotBody["active"])
	// Server expects secret nested under configuration
	cfg, ok := gotBody["configuration"].(map[string]any)
	require.True(t, ok, "configuration object expected on Server webhook create")
	assert.Equal(t, "redacted-test-value", cfg["secret"])
	assert.Equal(t, "99", hook.ID)
}

func TestServerClient_CreateWebhook_OmitsConfigurationWhenNoSecret(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"name":"bitbottle","url":"https://example.com/h","active":true,"events":["repo:refs_changed"]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateWebhook("PROJ", "repo", backend.CreateWebhookInput{
		URL: "https://example.com/h", Events: []string{"repo:refs_changed"}, Active: true,
	})
	require.NoError(t, err)
	_, hasCfg := gotBody["configuration"]
	assert.False(t, hasCfg, "configuration must be omitted when no secret is supplied")
}

func TestServerClient_DeleteWebhook_IssuesDelete(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeleteWebhook("PROJ", "repo", "42")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/projects/PROJ/repos/repo/webhooks/42", gotPath)
}
