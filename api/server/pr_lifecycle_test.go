package server_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_UpdatePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.UpdatePR("MYPROJ", "my-service", 42, backend.UpdatePRInput{Title: "New title"})
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42", gotPath)
}

func TestServerClient_UpdatePR_SendsBody(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.UpdatePR("MYPROJ", "my-service", 42, backend.UpdatePRInput{Title: "Updated", Description: "New body"})
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &payload))
	assert.Equal(t, "Updated", payload["title"])
	assert.Equal(t, "New body", payload["description"])
}

func TestServerClient_DeclinePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.DeclinePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/decline", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestServerClient_ReopenPR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var paths, methods []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		methods = append(methods, r.Method)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.ReopenPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	// First call GETs the PR (to read its current version), second call POSTs
	// to .../reopen.
	require.Len(t, paths, 2)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42", paths[0])
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/reopen", paths[1])
	assert.Equal(t, http.MethodPost, methods[1])
}

// TestServerClient_ReopenPR_SendsCurrentVersion verifies that ReopenPR fetches
// the PR first and threads its version into the reopen request body — the
// field Bitbucket Server's optimistic-concurrency layer requires to avoid
// HTTP 409 "out-of-date information" against any non-zero-version PR.
func TestServerClient_ReopenPR_SendsCurrentVersion(t *testing.T) {
	t.Parallel()
	var reopenBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			reopenBody, _ = io.ReadAll(r.Body)
		}
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.ReopenPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	// pr_get.json has "version":3; the reopen body must echo it back so the
	// server's optimistic-concurrency check passes.
	assert.Contains(t, string(reopenBody), `"version":3`)
}

func TestServerClient_ReopenPR_PropagatesError(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"PR not found"}]}`)
	})
	err := client.ReopenPR("MYPROJ", "my-service", 999)
	require.Error(t, err)
	// 404 must surface as the typed pr.not_found code so errfmt renders
	// the PR-specific envelope rather than a generic "not_found".
	assert.ErrorIs(t, err, backend.ErrNotFound)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de))
	assert.Equal(t, backend.CodePRNotFound, de.Code,
		"ReopenPR 404 must be stamped with pr.not_found")
	assert.Equal(t, "999", de.ID)
}

// TestServerClient_ReopenPR_Conflict_Returns409 verifies that a 409 from the
// /reopen endpoint (e.g. PR already open or merged-not-declined) surfaces as
// a typed *backend.DomainError with Kind=ErrConflict, and that the server's
// cause string is preserved into Message so renderers can show it.
func TestServerClient_ReopenPR_Conflict_Returns409(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		// First call (GET) returns the PR; second call (POST /reopen) 409s.
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `{"errors":[{"message":"Pull request is not declined"}]}`)
			return
		}
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.ReopenPR("MYPROJ", "my-service", 42)
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrConflict)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de), "expected *backend.DomainError")
	assert.Equal(t, backend.ErrConflict, de.Kind)
	assert.Contains(t, de.Message, "Pull request is not declined",
		"server cause string must be preserved into Message")
}

func TestServerClient_UnapprovePR_DeletesApproveEndpoint(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.UnapprovePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/approve", gotPath,
		"UnapprovePR must DELETE .../approve, not .../participants/~")
	assert.Equal(t, http.MethodDelete, gotMethod)
}

func TestServerClient_ReadyPR_GetsThenPutsFullBody(t *testing.T) {
	t.Parallel()
	var methods []string
	var bodies []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.ReadyPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	// First call GET, second call PUT with full body.
	require.Len(t, methods, 2)
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodPut, methods[1])
	put := bodies[1]
	assert.Contains(t, put, `"draft":false`)
	assert.Contains(t, put, `"title":"Fix login bug"`)
	assert.Contains(t, put, `"fromRef"`)
	assert.Contains(t, put, `"toRef"`)
}

func TestServerClient_UnreadyPR_GetsThenPutsFullBody(t *testing.T) {
	t.Parallel()
	var methods []string
	var bodies []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.UnreadyPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	// First call GET, second call PUT with full body.
	require.Len(t, methods, 2)
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodPut, methods[1])
	put := bodies[1]
	assert.Contains(t, put, `"draft":true`)
	assert.Contains(t, put, `"title":"Fix login bug"`)
	assert.Contains(t, put, `"fromRef"`)
	assert.Contains(t, put, `"toRef"`)
}

// TestServerClient_UpdatePR_IncludesVersion verifies that UpdatePR first GETs
// the current PR to read its version, then PUTs with that version in the body.
// Bitbucket Server rejects a PUT without version with HTTP 400.
func TestServerClient_UpdatePR_IncludesVersion(t *testing.T) {
	t.Parallel()
	var putBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			putBody, _ = io.ReadAll(r.Body)
			// Return a plausible PR response with version bumped.
			_, _ = io.WriteString(w, `{"id":135,"version":8,"title":"new","description":"newdesc","state":"OPEN","draft":false,"author":{"user":{"slug":"alice","displayName":"Alice"}},"fromRef":{"id":"refs/heads/feat/x","displayId":"feat/x"},"toRef":{"id":"refs/heads/main","displayId":"main"},"links":{"self":[{"href":"https://bb.example.com/pull-requests/135"}]}}`)
			return
		}
		// GET: return PR with version 7.
		_, _ = io.WriteString(w, `{"id":135,"version":7,"title":"old","description":"old","state":"OPEN","draft":false,"author":{"user":{"slug":"alice","displayName":"Alice"}},"fromRef":{"id":"refs/heads/feat/x","displayId":"feat/x"},"toRef":{"id":"refs/heads/main","displayId":"main"},"links":{"self":[{"href":"https://bb.example.com/pull-requests/135"}]}}`)
	})
	_, err := client.UpdatePR("ns", "slug", 135, backend.UpdatePRInput{Title: "new", Description: "newdesc"})
	require.NoError(t, err)
	// PUT body must include the version fetched from GET (7).
	assert.Contains(t, string(putBody), `"version":7`,
		"UpdatePR PUT body must include the current version for Server optimistic concurrency")
}

func TestServerClient_RemoveReviewers_DeletesPerUser(t *testing.T) {
	t.Parallel()
	var paths, methods []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		methods = append(methods, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.RemoveReviewers("MYPROJ", "my-service", 42, []string{"alice", "bob"})
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/participants/alice", paths[0])
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/participants/bob", paths[1])
	assert.Equal(t, http.MethodDelete, methods[0])
	assert.Equal(t, http.MethodDelete, methods[1])
}

func TestServerClient_RemoveReviewers_404IsIgnored(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"not found"}]}`)
	})
	// 404 for a non-participant should not be an error
	err := client.RemoveReviewers("MYPROJ", "my-service", 42, []string{"ghost"})
	require.NoError(t, err)
}

func TestServerClient_RequestReview_GetsAndPutsPR(t *testing.T) {
	t.Parallel()
	var methods []string
	var bodies []string
	callCount := 0
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		callCount++
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.RequestReview("MYPROJ", "my-service", 42, []string{"alice", "bob"})
	require.NoError(t, err)
	// First call GET, second call PUT
	require.Len(t, methods, 2)
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodPut, methods[1])
	// PUT body should contain reviewers
	assert.Contains(t, bodies[1], "alice")
	assert.Contains(t, bodies[1], "bob")
}
