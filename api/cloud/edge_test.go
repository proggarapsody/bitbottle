package cloud_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloudClient_ListRepos_FirstPageWithNextLink verifies that when the
// server returns a response that includes a "next" cursor, the current
// implementation still decodes and returns the values from the first page
// without following the next link. This documents the current single-page
// behaviour of the cloud adapter.
func TestCloudClient_ListRepos_FirstPageWithNextLink(t *testing.T) {
	t.Parallel()
	var requestCount int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		// First-page body containing a "next" link.
		body := `{"pagelen":2,"page":1,"size":4,"values":[
			{"type":"repository","full_name":"ws/repo-a","slug":"repo-a","name":"repo-a","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/repo-a"}}},
			{"type":"repository","full_name":"ws/repo-b","slug":"repo-b","name":"repo-b","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/repo-b"}}}
		],"next":"https://api.bitbucket.org/2.0/repositories?page=2"}`
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	repos, err := client.ListRepos(2)
	require.NoError(t, err)
	require.Len(t, repos, 2, "only first page values are returned")
	assert.Equal(t, "repo-a", repos[0].Slug)
	assert.Equal(t, "repo-b", repos[1].Slug)
	// The client should not transparently chase the cursor.
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount),
		"pagination cursor must not be auto-followed")
}

// TestCloudClient_ListPRs_FirstPageWithNextLink verifies the same
// single-page behaviour for pull-request listings.
func TestCloudClient_ListPRs_FirstPageWithNextLink(t *testing.T) {
	t.Parallel()
	var requestCount int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		body := `{"pagelen":2,"page":1,"size":4,"values":[
			{"type":"pullrequest","id":1,"title":"PR 1","state":"OPEN","author":{"display_name":"Alice","account_id":"alice"},"source":{"branch":{"name":"feat/a"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":"https://bitbucket.org/ws/repo/pull-requests/1"}}},
			{"type":"pullrequest","id":2,"title":"PR 2","state":"OPEN","author":{"display_name":"Alice","account_id":"alice"},"source":{"branch":{"name":"feat/b"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":"https://bitbucket.org/ws/repo/pull-requests/2"}}}
		],"next":"https://api.bitbucket.org/2.0/repositories/ws/repo/pullrequests?page=2"}`
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	prs, err := client.ListPRs("ws", "repo", "OPEN", 2)
	require.NoError(t, err)
	require.Len(t, prs, 2)
	assert.Equal(t, 1, prs[0].ID)
	assert.Equal(t, 2, prs[1].ID)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount),
		"pagination cursor must not be auto-followed")
}

// TestCloudClient_ListRepos_MalformedNextDoesNotPanic verifies that even when
// the "next" field contains a non-parseable URL the client decodes the first
// page without panicking and returns the values that were collected.
func TestCloudClient_ListRepos_MalformedNextDoesNotPanic(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body := `{"pagelen":1,"page":1,"size":2,"values":[
			{"type":"repository","full_name":"ws/only","slug":"only","name":"only","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/only"}}}
		],"next":"::::not-a-url::::"}`
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	// Must not panic.
	require.NotPanics(t, func() {
		repos, err := client.ListRepos(1)
		require.NoError(t, err)
		require.Len(t, repos, 1)
		assert.Equal(t, "only", repos[0].Slug)
	})
}

// TestCloudClient_BearerAuth_Precedence verifies that a non-empty token wins
// over a non-empty username (Bearer is preferred).
func TestCloudClient_BearerAuth_Precedence(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "the-token", "alice")
	_, _ = client.GetCurrentUser()
	assert.Equal(t, "Bearer the-token", gotAuth)
}

// TestCloudClient_BasicAuth_WhenTokenEmpty verifies that when the token is
// empty and a username is supplied the client sends a Basic auth header.
func TestCloudClient_BasicAuth_WhenTokenEmpty(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "", "alice")
	_, _ = client.GetCurrentUser()
	require.True(t, strings.HasPrefix(gotAuth, "Basic "),
		"expected Basic auth header, got %q", gotAuth)
}

// TestCloudClient_NoAuth_WhenBothEmpty verifies no Authorization header is
// attached when neither a token nor a username is supplied.
func TestCloudClient_NoAuth_WhenBothEmpty(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "", "")
	_, _ = client.GetCurrentUser()
	assert.Empty(t, gotAuth)
}

