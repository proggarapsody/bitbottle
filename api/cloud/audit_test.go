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

// ── ListAuditLog ──────────────────────────────────────────────────────────────

func TestCloudClient_ListAuditLog_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workspaces/myworkspace/log/audit", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{
				"actor":{"account_id":"aid-1","display_name":"Alice","nickname":"alice"},
				"action":"workspace.member.create",
				"object":{"type":"team","name":"acme"},
				"created_on":"2024-01-15T10:00:00.000000+00:00"
			}
		]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	events, err := client.ListAuditLog("myworkspace", backend.AuditLogOpts{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	e := events[0]
	assert.Equal(t, "aid-1", e.Actor.AccountID)
	assert.Equal(t, "Alice", e.Actor.DisplayName)
	assert.Equal(t, "alice", e.Actor.NickName)
	assert.Equal(t, "workspace.member.create", e.Action)
	assert.Equal(t, "team", e.Object.Type)
	assert.Equal(t, "acme", e.Object.Name)
	assert.Equal(t, "2024-01-15T10:00:00.000000+00:00", e.CreatedAt)
}

func TestCloudClient_ListAuditLog_WithActionFilter(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListAuditLog("myworkspace", backend.AuditLogOpts{Action: "workspace.member.create"})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "action=workspace.member.create")
}

func TestCloudClient_ListAuditLog_HTTP403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Access denied"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListAuditLog("myworkspace", backend.AuditLogOpts{})
	require.Error(t, err)
}
