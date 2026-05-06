package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListWebhooks_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWebhooks("ws", "repo")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/hooks", gotPath)
}

func TestCloudClient_ListWebhooks_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
            {"uuid":"{abc-1}","url":"https://example.com/hook","active":true,"events":["repo:push","pullrequest:created"]},
            {"uuid":"{abc-2}","url":"https://other.example/hook","active":false,"events":["repo:push"]}
        ]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	hooks, err := client.ListWebhooks("ws", "repo")
	require.NoError(t, err)
	require.Len(t, hooks, 2)
	assert.Equal(t, "abc-1", hooks[0].ID)
	assert.Equal(t, "https://example.com/hook", hooks[0].URL)
	assert.True(t, hooks[0].Active)
	assert.Equal(t, []string{"repo:push", "pullrequest:created"}, hooks[0].Events)
	assert.Equal(t, "abc-2", hooks[1].ID)
	assert.False(t, hooks[1].Active)
}

func TestCloudClient_GetWebhook_BracesUUIDInPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{abc-1}","url":"https://example.com/hook","active":true,"events":["repo:push"]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	hook, err := client.GetWebhook("ws", "repo", "abc-1")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/hooks/{abc-1}", gotPath)
	assert.Equal(t, "abc-1", hook.ID)
	assert.Equal(t, "https://example.com/hook", hook.URL)
}

func TestCloudClient_CreateWebhook_PostsExpectedBody(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{new-1}","url":"https://example.com/hook","active":true,"events":["repo:push"]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	hook, err := client.CreateWebhook("ws", "repo", backend.CreateWebhookInput{
		URL:    "https://example.com/hook",
		Events: []string{"repo:push"},
		Active: true,
		Secret: "shhh",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/ws/repo/hooks", gotPath)
	assert.Equal(t, "https://example.com/hook", gotBody["url"])
	assert.Equal(t, true, gotBody["active"])
	assert.Equal(t, "shhh", gotBody["secret"])
	// Cloud calls events "events" and uses a description="bitbottle" placeholder if none given.
	assert.Contains(t, gotBody, "events")
	assert.Equal(t, "new-1", hook.ID)
}

func TestCloudClient_CreateWebhook_OmitsSecretWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{x}","url":"https://example.com/hook","active":true,"events":["repo:push"]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateWebhook("ws", "repo", backend.CreateWebhookInput{
		URL: "https://example.com/hook", Events: []string{"repo:push"}, Active: true,
	})
	require.NoError(t, err)
	_, hasSecret := gotBody["secret"]
	assert.False(t, hasSecret, "secret must be omitted from body when empty so Cloud doesn't store an empty string")
}

func TestCloudClient_DeleteWebhook_BracesUUIDInPath(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeleteWebhook("ws", "repo", "abc-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/ws/repo/hooks/{abc-1}", gotPath)
}
