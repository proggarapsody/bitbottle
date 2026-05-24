package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func pagedGroups(names ...string) map[string]any {
	values := make([]any, len(names))
	for i, n := range names {
		values[i] = map[string]any{"name": n}
	}
	return map[string]any{
		"values": values, "isLastPage": true, "size": len(names),
	}
}

func TestServerClient_ListGroups(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedGroups("developers", "admins"))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	groups, err := c.ListGroups("", 25)
	require.NoError(t, err)
	assert.Contains(t, seenPath, "/admin/groups")
	assert.Equal(t, []backend.Group{{Name: "developers"}, {Name: "admins"}}, groups)
}

func TestServerClient_ListGroups_WithFilter(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedGroups("devs"))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	groups, err := c.ListGroups("dev", 25)
	require.NoError(t, err)
	assert.Contains(t, seenPath, "filter=dev")
	assert.Len(t, groups, 1)
}

func TestServerClient_CreateGroup(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "newgroup"})
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	g, err := c.CreateGroup("newgroup")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Equal(t, "/admin/groups", seenPath)
	assert.Equal(t, "newgroup", seenBody["name"])
	assert.Equal(t, backend.Group{Name: "newgroup"}, g)
}

func TestServerClient_DeleteGroup(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.RequestURI()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	err := c.DeleteGroup("oldgroup")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Contains(t, seenPath, "name=oldgroup")
}

func TestServerClient_ListGroupMembers(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"values": []any{
				map[string]any{"name": "alice", "displayName": "Alice", "emailAddress": "alice@example.com"},
				map[string]any{"name": "bob", "displayName": "Bob", "emailAddress": "bob@example.com"},
			},
			"isLastPage": true,
		})
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	members, err := c.ListGroupMembers("developers", 25)
	require.NoError(t, err)
	assert.Contains(t, seenPath, "/admin/groups/more-members")
	assert.Contains(t, seenPath, "context=developers")
	assert.Equal(t, []backend.GroupMember{
		{Name: "alice", DisplayName: "Alice", EmailAddress: "alice@example.com"},
		{Name: "bob", DisplayName: "Bob", EmailAddress: "bob@example.com"},
	}, members)
}

func TestServerClient_AddGroupMember(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	err := c.AddGroupMember("developers", "alice")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Equal(t, "/admin/users/add-group", seenPath)
	assert.Equal(t, "alice", seenBody["user"])
	assert.Equal(t, "developers", seenBody["group"])
}

func TestServerClient_RemoveGroupMember(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	err := c.RemoveGroupMember("developers", "alice")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Equal(t, "/admin/users/remove-group", seenPath)
	assert.Equal(t, "alice", seenBody["user"])
	assert.Equal(t, "developers", seenBody["group"])
}

func TestServer_Group_ImplementsInterfaces(t *testing.T) {
	t.Parallel()
	var _ backend.GroupClient = (*server.Client)(nil)
	var _ backend.GroupMemberClient = (*server.Client)(nil)
}
