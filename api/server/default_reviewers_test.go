package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

// repoGetJSON mimics the BBS repo response with a numeric ID.
const repoGetJSON = `{
  "id": 17,
  "slug": "my-service",
  "name": "my-service",
  "project": {"key": "MYPROJ", "name": "MyProj"},
  "scmId": "git",
  "links": {"self": [{"href": "https://bb.example.com/projects/MYPROJ/repos/my-service"}]}
}`

const reviewersJSON = `[
  {"name": "alice", "displayName": "Alice"},
  {"name": "bob",   "displayName": "Bob"}
]`

func TestServerClient_DefaultReviewers_QueryParamsAndPath(t *testing.T) {
	t.Parallel()
	var seenPaths []string
	var lastQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPaths = append(seenPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		// Repo lookup for ID resolution.
		if r.URL.Path == "/projects/MYPROJ/repos/my-service" {
			_, _ = w.Write([]byte(repoGetJSON))
			return
		}
		// The default-reviewers endpoint.
		if r.URL.Path == "/rest/default-reviewers/1.0/projects/MYPROJ/repos/my-service/reviewers" {
			lastQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(reviewersJSON))
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.DefaultReviewers("MYPROJ", "my-service", "feat/x", "main")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "alice", got[0].Slug)
	assert.Equal(t, "Alice", got[0].DisplayName)
	assert.Equal(t, "bob", got[1].Slug)

	// Path traffic: GetRepo via the regular transport, then default-reviewers
	// on the dedicated /rest/default-reviewers/1.0 transport.
	assert.Equal(t, []string{
		"/projects/MYPROJ/repos/my-service",
		"/rest/default-reviewers/1.0/projects/MYPROJ/repos/my-service/reviewers",
	}, seenPaths)

	// Query params: source/target IDs both come from the GetRepo result; ref
	// IDs are ensureRefsHeads-expanded (refs/heads/<branch>).
	assert.Contains(t, lastQuery, "sourceRepoId=17")
	assert.Contains(t, lastQuery, "targetRepoId=17")
	assert.Contains(t, lastQuery, "sourceRefId=refs%2Fheads%2Ffeat%2Fx")
	assert.Contains(t, lastQuery, "targetRefId=refs%2Fheads%2Fmain")
}

func TestServerClient_DefaultReviewers_RepoLookupFails(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	_, err := client.DefaultReviewers("MYPROJ", "missing", "feat/x", "main")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default reviewers")
}

func TestServerClient_DefaultReviewers_EmptyResponse(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/projects/MYPROJ/repos/my-service" {
			_, _ = w.Write([]byte(repoGetJSON))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.DefaultReviewers("MYPROJ", "my-service", "feat/x", "main")
	require.NoError(t, err)
	assert.Empty(t, got, "no configured defaults must produce an empty (not nil) slice")
}
