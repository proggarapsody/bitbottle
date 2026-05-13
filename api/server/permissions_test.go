package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// ── fixtures ──────────────────────────────────────────────────────────────────

const projectUsersJSON = `{
  "values": [
    {"user":{"slug":"alice","displayName":"Alice A"},"permission":"PROJECT_ADMIN"},
    {"user":{"slug":"bob","displayName":"Bob B"},"permission":"PROJECT_READ"}
  ],
  "isLastPage": true
}`

const projectGroupsJSON = `{
  "values": [
    {"group":{"name":"devs"},"permission":"PROJECT_WRITE"}
  ],
  "isLastPage": true
}`

const repoUsersJSON = `{
  "values": [
    {"user":{"slug":"carol","displayName":"Carol C"},"permission":"REPO_ADMIN"}
  ],
  "isLastPage": true
}`

const repoGroupsJSON = `{
  "values": [
    {"group":{"name":"qa team"},"permission":"REPO_READ"}
  ],
  "isLastPage": true
}`

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestServer(t *testing.T, handler http.HandlerFunc) (*server.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "alice"), srv
}

// ── ListProjectPermissions ────────────────────────────────────────────────────

func TestServerClient_ListProjectPermissions_MergesAndSorts(t *testing.T) {
	t.Parallel()
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/projects/MYPROJ/permissions/users" {
			_, _ = w.Write([]byte(projectUsersJSON))
		} else {
			_, _ = w.Write([]byte(projectGroupsJSON))
		}
	}))

	got, err := client.ListProjectPermissions(context.Background(), "MYPROJ")
	require.NoError(t, err)
	require.Len(t, got, 3)

	// sorted: ADMIN first, then WRITE, then READ
	assert.Equal(t, "PROJECT_ADMIN", got[0].Permission)
	assert.Equal(t, "alice", got[0].Subject.Slug)
	assert.Equal(t, "user", got[0].Subject.Kind)
	assert.Equal(t, "Alice A", got[0].Subject.DisplayName)

	assert.Equal(t, "PROJECT_WRITE", got[1].Permission)
	assert.Equal(t, "group", got[1].Subject.Kind)
	assert.Equal(t, "devs", got[1].Subject.Name)

	assert.Equal(t, "PROJECT_READ", got[2].Permission)
	assert.Equal(t, "bob", got[2].Subject.Slug)
}

// ── GrantProjectPermission ────────────────────────────────────────────────────

func TestServerClient_GrantProjectPermission_User(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenQuery url.Values
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.GrantProjectPermission(context.Background(), "MYPROJ",
		backend.PermissionSubject{Kind: "user", Slug: "alice"}, "PROJECT_WRITE")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, seenMethod)
	assert.Equal(t, "/projects/MYPROJ/permissions/users", seenPath)
	assert.Equal(t, "alice", seenQuery.Get("name"))
	assert.Equal(t, "PROJECT_WRITE", seenQuery.Get("permission"))
}

func TestServerClient_GrantProjectPermission_Group(t *testing.T) {
	t.Parallel()
	var seenPath string
	var seenQuery url.Values
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.GrantProjectPermission(context.Background(), "MYPROJ",
		backend.PermissionSubject{Kind: "group", Name: "qa team"}, "PROJECT_READ")
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/permissions/groups", seenPath)
	assert.Equal(t, "qa team", seenQuery.Get("name"))
}

// ── RevokeProjectPermission ───────────────────────────────────────────────────

func TestServerClient_RevokeProjectPermission_User(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenQuery url.Values
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.RevokeProjectPermission(context.Background(), "MYPROJ",
		backend.PermissionSubject{Kind: "user", Slug: "bob"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Equal(t, "/projects/MYPROJ/permissions/users", seenPath)
	assert.Equal(t, "bob", seenQuery.Get("name"))
}

// ── ListRepoPermissions ───────────────────────────────────────────────────────

func TestServerClient_ListRepoPermissions_MergesAndSorts(t *testing.T) {
	t.Parallel()
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/projects/MYPROJ/repos/my-repo/permissions/users" {
			_, _ = w.Write([]byte(repoUsersJSON))
		} else {
			_, _ = w.Write([]byte(repoGroupsJSON))
		}
	}))

	got, err := client.ListRepoPermissions(context.Background(), "MYPROJ", "my-repo")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "REPO_ADMIN", got[0].Permission)
	assert.Equal(t, "carol", got[0].Subject.Slug)
	assert.Equal(t, "REPO_READ", got[1].Permission)
	assert.Equal(t, "qa team", got[1].Subject.Name)
}

// ── GrantRepoPermission ───────────────────────────────────────────────────────

func TestServerClient_GrantRepoPermission_User(t *testing.T) {
	t.Parallel()
	var seenPath string
	var seenQuery url.Values
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.GrantRepoPermission(context.Background(), "MYPROJ", "my-repo",
		backend.PermissionSubject{Kind: "user", Slug: "carol"}, "REPO_WRITE")
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-repo/permissions/users", seenPath)
	assert.Equal(t, "carol", seenQuery.Get("name"))
	assert.Equal(t, "REPO_WRITE", seenQuery.Get("permission"))
}

// ── RevokeRepoPermission ──────────────────────────────────────────────────────

func TestServerClient_RevokeRepoPermission_Group(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenQuery url.Values
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		seenQuery = r.URL.Query()
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.RevokeRepoPermission(context.Background(), "MYPROJ", "my-repo",
		backend.PermissionSubject{Kind: "group", Name: "qa team"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Equal(t, "/projects/MYPROJ/repos/my-repo/permissions/groups", seenPath)
	assert.Equal(t, "qa team", seenQuery.Get("name"))
}

// ── group name URL encoding ───────────────────────────────────────────────────

func TestServerClient_GroupNameWithSpaces_URLEncoded(t *testing.T) {
	t.Parallel()
	var seenRawQuery string
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.GrantProjectPermission(context.Background(), "MYPROJ",
		backend.PermissionSubject{Kind: "group", Name: "my group"}, "PROJECT_READ")
	require.NoError(t, err)
	// url.Values.Encode percent-encodes spaces as +; the net/http server
	// decodes + back to space so the server sees "my group".  Verify the
	// raw query is not sending a literal space.
	assert.NotContains(t, seenRawQuery, " ", "raw query must not contain unencoded spaces")
}

// ── paged response ────────────────────────────────────────────────────────────

func TestServerClient_ListProjectPermissions_Paged(t *testing.T) {
	t.Parallel()
	pages := map[string][]string{
		"/projects/MYPROJ/permissions/users": {
			`{"values":[{"user":{"slug":"u1","displayName":"U1"},"permission":"PROJECT_READ"}],"isLastPage":false,"nextPageStart":1}`,
			`{"values":[{"user":{"slug":"u2","displayName":"U2"},"permission":"PROJECT_READ"}],"isLastPage":true}`,
		},
		"/projects/MYPROJ/permissions/groups": {
			`{"values":[],"isLastPage":true}`,
		},
	}
	hits := map[string]int{}
	client, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := pages[r.URL.Path]
		_, _ = w.Write([]byte(p[hits[r.URL.Path]]))
		hits[r.URL.Path]++
	}))

	got, err := client.ListProjectPermissions(context.Background(), "MYPROJ")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "u1", got[0].Subject.Slug)
	assert.Equal(t, "u2", got[1].Subject.Slug)
}
