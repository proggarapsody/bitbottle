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

const wsProjectUserPermsJSON = `{
  "values": [
    {"permission": "write", "user": {"account_id": "abc123", "display_name": "Alice", "nickname": "alice"}},
    {"permission": "admin", "user": {"account_id": "def456", "display_name": "Bob",   "nickname": "bob"}}
  ]
}`

const wsProjectGroupPermsJSON = `{
  "values": [
    {"permission": "read", "group": {"slug": "devs", "name": "Developers"}}
  ]
}`

func TestCloudClient_ListWorkspaceProjectPerms_UserPath(t *testing.T) {
	t.Parallel()
	var gotPaths []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/workspaces/myws/projects/PROJ/permissions/users" {
			_, _ = w.Write([]byte(wsProjectUserPermsJSON))
		} else {
			_, _ = w.Write([]byte(`{"values":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListWorkspaceProjectPerms("myws", "PROJ", 0)
	require.NoError(t, err)
	assert.Contains(t, gotPaths, "/workspaces/myws/projects/PROJ/permissions/users")
	assert.Contains(t, gotPaths, "/workspaces/myws/projects/PROJ/permissions/groups")
	// Two user entries + 0 group entries
	require.Len(t, got, 2)
	assert.Equal(t, "write", got[0].Permission)
	require.NotNil(t, got[0].User)
	assert.Equal(t, "alice", got[0].User.Slug)
	assert.Equal(t, "Alice", got[0].User.DisplayName)
	assert.Nil(t, got[0].Group)
}

func TestCloudClient_ListWorkspaceProjectPerms_GroupPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/workspaces/myws/projects/PROJ/permissions/groups" {
			_, _ = w.Write([]byte(wsProjectGroupPermsJSON))
		} else {
			_, _ = w.Write([]byte(`{"values":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListWorkspaceProjectPerms("myws", "PROJ", 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "read", got[0].Permission)
	require.NotNil(t, got[0].Group)
	assert.Equal(t, "devs", got[0].Group.Slug)
	assert.Equal(t, "Developers", got[0].Group.Name)
	assert.Nil(t, got[0].User)
}

func TestCloudClient_ListWorkspaceProjectPerms_MergesBoth(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/workspaces/myws/projects/PROJ/permissions/users":
			_, _ = w.Write([]byte(wsProjectUserPermsJSON))
		case "/workspaces/myws/projects/PROJ/permissions/groups":
			_, _ = w.Write([]byte(wsProjectGroupPermsJSON))
		default:
			_, _ = w.Write([]byte(`{"values":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListWorkspaceProjectPerms("myws", "PROJ", 0)
	require.NoError(t, err)
	require.Len(t, got, 3) // 2 users + 1 group
}

func TestCloudClient_GrantWorkspaceProjectPerm_User(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.GrantWorkspaceProjectPerm("myws", "PROJ", backend.WorkspaceProjectPermInput{
		Permission: "write",
		UserSlug:   "alice",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/permissions/users/alice", gotPath)
	assert.Equal(t, "write", gotBody["permission"])
}

func TestCloudClient_GrantWorkspaceProjectPerm_Group(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.GrantWorkspaceProjectPerm("myws", "PROJ", backend.WorkspaceProjectPermInput{
		Permission: "read",
		GroupSlug:  "devs",
	})
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/permissions/groups/devs", gotPath)
}

func TestCloudClient_RevokeWorkspaceProjectPerm_User(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.RevokeWorkspaceProjectPerm("myws", "PROJ", "alice", false)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/permissions/users/alice", gotPath)
}

func TestCloudClient_RevokeWorkspaceProjectPerm_Group(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.RevokeWorkspaceProjectPerm("myws", "PROJ", "devs", true)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/permissions/groups/devs", gotPath)
}

func TestCloudClient_ListWorkspaceProjectPerms_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.ListWorkspaceProjectPerms("myws", "PROJ", 0)
	require.Error(t, err)
}
