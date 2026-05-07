package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

const restrictionsJSON = `{
  "values": [
    {
      "id": 1,
      "type": "fast-forward-only",
      "matcher": {
        "id": "main",
        "displayId": "main",
        "type": {"id": "BRANCH", "name": "Branch"},
        "active": true
      },
      "users": [{"name": "alice", "displayName": "Alice"}],
      "groups": ["devs"]
    },
    {
      "id": 2,
      "type": "no-deletes",
      "matcher": {
        "id": "release/*",
        "displayId": "release/*",
        "type": {"id": "PATTERN", "name": "Pattern"},
        "active": true
      },
      "users": [],
      "groups": []
    }
  ],
  "size": 2,
  "isLastPage": true,
  "start": 0
}`

// TestServerClient_ListBranchProtections verifies the wire path and the
// projection from BBS's nested matcher shape into the flat domain type.
func TestServerClient_ListBranchProtections(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(restrictionsJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.ListBranchProtections("MYPROJ", "my-service", 25)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "/rest/branch-permissions/2.0/projects/MYPROJ/repos/my-service/restrictions", seenPath)

	assert.Equal(t, 1, got[0].ID)
	assert.Equal(t, "fast-forward-only", got[0].Type)
	assert.Equal(t, "main", got[0].MatcherID)
	assert.Equal(t, "BRANCH", got[0].MatcherKind)
	assert.Equal(t, []string{"alice"}, got[0].Users)
	assert.Equal(t, []string{"devs"}, got[0].Groups)

	assert.Equal(t, "release/*", got[1].MatcherID)
	assert.Equal(t, "PATTERN", got[1].MatcherKind)
	assert.Empty(t, got[1].Users)
	assert.Empty(t, got[1].Groups)
}

// TestServerClient_CreateBranchProtection covers the inverse: a flat domain
// input must round-trip into BBS's nested {matcher:{id,type:{id}}} payload,
// and the response is decoded back into a domain type.
func TestServerClient_CreateBranchProtection(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	var seenPath, seenContentType string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
          "id": 42,
          "type": "fast-forward-only",
          "matcher": {"id":"main","displayId":"main","type":{"id":"BRANCH"},"active":true},
          "users": [{"name":"alice"}],
          "groups": ["devs"]
        }`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.CreateBranchProtection("MYPROJ", "my-service", backend.CreateBranchProtectionInput{
		Type:        "fast-forward-only",
		MatcherID:   "main",
		MatcherKind: "BRANCH",
		Users:       []string{"alice"},
		Groups:      []string{"devs"},
	})
	require.NoError(t, err)

	assert.Equal(t, "/rest/branch-permissions/2.0/projects/MYPROJ/repos/my-service/restrictions", seenPath)
	assert.True(t, strings.HasPrefix(seenContentType, "application/json"),
		"expected JSON Content-Type, got %q", seenContentType)

	var sent struct {
		Type    string `json:"type"`
		Matcher struct {
			ID   string `json:"id"`
			Type struct {
				ID string `json:"id"`
			} `json:"type"`
		} `json:"matcher"`
		Users  []string `json:"users"`
		Groups []string `json:"groups"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "fast-forward-only", sent.Type)
	assert.Equal(t, "main", sent.Matcher.ID)
	assert.Equal(t, "BRANCH", sent.Matcher.Type.ID)
	assert.Equal(t, []string{"alice"}, sent.Users)
	assert.Equal(t, []string{"devs"}, sent.Groups)

	assert.Equal(t, 42, got.ID)
	assert.Equal(t, "fast-forward-only", got.Type)
	assert.Equal(t, "main", got.MatcherID)
	assert.Equal(t, "BRANCH", got.MatcherKind)
}

// TestServerClient_CreateBranchProtection_DefaultMatcherKind: empty
// MatcherKind defaults to "BRANCH" — the most common case for users typing
// `--branch main` on the CLI.
func TestServerClient_CreateBranchProtection_DefaultMatcherKind(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"type":"no-deletes","matcher":{"id":"main","type":{"id":"BRANCH"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	_, err := client.CreateBranchProtection("MYPROJ", "my-service", backend.CreateBranchProtectionInput{
		Type:      "no-deletes",
		MatcherID: "main",
	})
	require.NoError(t, err)

	var sent struct {
		Matcher struct {
			Type struct {
				ID string `json:"id"`
			} `json:"type"`
		} `json:"matcher"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "BRANCH", sent.Matcher.Type.ID, "empty MatcherKind must default to BRANCH")
}

// TestServerClient_DeleteBranchProtection: numeric ID into the path; 204
// is the expected success status.
func TestServerClient_DeleteBranchProtection(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	require.NoError(t, client.DeleteBranchProtection("MYPROJ", "my-service", 42))
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Equal(t, "/rest/branch-permissions/2.0/projects/MYPROJ/repos/my-service/restrictions/42", seenPath)
}

// TestServerClient_ListBranchProtections_Paged: the dedicated transport
// must use the Server paginator so multi-page responses concatenate.
func TestServerClient_ListBranchProtections_Paged(t *testing.T) {
	t.Parallel()
	pages := []string{
		`{"values":[{"id":1,"type":"no-deletes","matcher":{"id":"main","type":{"id":"BRANCH"}}}],"size":1,"isLastPage":false,"nextPageStart":1,"start":0}`,
		`{"values":[{"id":2,"type":"read-only","matcher":{"id":"release/*","type":{"id":"PATTERN"}}}],"size":1,"isLastPage":true,"start":1}`,
	}
	var hits int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pages[hits]))
		hits++
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.ListBranchProtections("MYPROJ", "my-service", 100)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].ID)
	assert.Equal(t, 2, got[1].ID)
	assert.Equal(t, 2, hits, "expected exactly 2 page fetches")
}
