package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// ── ListIPAllowlists ──────────────────────────────────────────────────────────

func TestCloudClient_ListIPAllowlists_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workspaces/myworkspace/ipallowlists", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{
				"uuid":"{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}",
				"cidr":"10.0.0.0/8",
				"description":"Corporate VPN",
				"enabled":true
			}
		]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	entries, err := client.ListIPAllowlists("myworkspace")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}", e.UUID)
	assert.Equal(t, "10.0.0.0/8", e.CIDR)
	assert.Equal(t, "Corporate VPN", e.Description)
	assert.True(t, e.Enabled)
}

func TestCloudClient_ListIPAllowlists_Empty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	entries, err := client.ListIPAllowlists("myworkspace")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestCloudClient_ListIPAllowlists_HTTP403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Access denied"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListIPAllowlists("myworkspace")
	require.Error(t, err)
}

// ── CreateIPAllowlist ─────────────────────────────────────────────────────────

func TestCloudClient_CreateIPAllowlist_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/workspaces/myworkspace/ipallowlists", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid":"{new-uuid-1234}",
			"cidr":"192.168.1.0/24",
			"description":"Office",
			"enabled":true
		}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	entry, err := client.CreateIPAllowlist("myworkspace", backend.CreateIPAllowlistInput{
		CIDR:        "192.168.1.0/24",
		Description: "Office",
		Enabled:     true,
	})
	require.NoError(t, err)
	assert.Equal(t, "{new-uuid-1234}", entry.UUID)
	assert.Equal(t, "192.168.1.0/24", entry.CIDR)
	assert.Equal(t, "Office", entry.Description)
	assert.True(t, entry.Enabled)
}

func TestCloudClient_CreateIPAllowlist_HTTP400(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Invalid CIDR"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateIPAllowlist("myworkspace", backend.CreateIPAllowlistInput{
		CIDR: "not-a-cidr",
	})
	require.Error(t, err)
}

// ── DeleteIPAllowlist ─────────────────────────────────────────────────────────

func TestCloudClient_DeleteIPAllowlist_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/workspaces/myworkspace/ipallowlists/entry-uuid", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeleteIPAllowlist("myworkspace", "entry-uuid")
	require.NoError(t, err)
}

func TestCloudClient_DeleteIPAllowlist_HTTP404(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Not found"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeleteIPAllowlist("myworkspace", "nonexistent-uuid")
	require.Error(t, err)
}
