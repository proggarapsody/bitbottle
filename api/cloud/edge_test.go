package cloud_test

import (
	"fmt"
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

// TestCloudClient_ListRepos_FollowsNextLink verifies that when the server
// returns a paged response with a "next" URL the client follows it and
// accumulates values from all pages.
func TestCloudClient_ListRepos_FollowsNextLink(t *testing.T) {
	t.Parallel()
	var repoRequestCount int32
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/user" {
			_, _ = io.WriteString(w, `{"nickname":"ws","account_id":"ws","display_name":"WS"}`)
			return
		}
		count := atomic.AddInt32(&repoRequestCount, 1)
		if count == 1 {
			nextURL := srv.URL + "/repositories/ws?page=2"
			body := fmt.Sprintf(`{"pagelen":1,"page":1,"size":2,"values":[
				{"type":"repository","full_name":"ws/repo-a","slug":"repo-a","name":"repo-a","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/repo-a"}}}
			],"next":"%s"}`, nextURL)
			_, _ = io.WriteString(w, body)
		} else {
			body := `{"pagelen":1,"page":2,"size":2,"values":[
				{"type":"repository","full_name":"ws/repo-b","slug":"repo-b","name":"repo-b","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/repo-b"}}}
			],"next":""}`
			_, _ = io.WriteString(w, body)
		}
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	repos, err := client.ListRepos(1)
	require.NoError(t, err)
	require.Len(t, repos, 2, "all pages must be accumulated")
	assert.Equal(t, "repo-a", repos[0].Slug)
	assert.Equal(t, "repo-b", repos[1].Slug)
	assert.Equal(t, int32(2), atomic.LoadInt32(&repoRequestCount),
		"both repo pages must be fetched")
}

// TestCloudClient_ListPRs_FollowsNextLink verifies that ListPRs follows the
// pagination cursor across all available pages.
func TestCloudClient_ListPRs_FollowsNextLink(t *testing.T) {
	t.Parallel()
	var requestCount int32
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			nextURL := srv.URL + "/repositories/ws/repo/pullrequests?page=2"
			body := fmt.Sprintf(`{"pagelen":1,"page":1,"size":2,"values":[
				{"type":"pullrequest","id":1,"title":"PR 1","state":"OPEN","author":{"display_name":"Alice","account_id":"alice"},"source":{"branch":{"name":"feat/a"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":"https://bitbucket.org/ws/repo/pull-requests/1"}}}
			],"next":"%s"}`, nextURL)
			_, _ = io.WriteString(w, body)
		} else {
			body := `{"pagelen":1,"page":2,"size":2,"values":[
				{"type":"pullrequest","id":2,"title":"PR 2","state":"OPEN","author":{"display_name":"Alice","account_id":"alice"},"source":{"branch":{"name":"feat/b"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":"https://bitbucket.org/ws/repo/pull-requests/2"}}}
			],"next":""}`
			_, _ = io.WriteString(w, body)
		}
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	prs, err := client.ListPRs("ws", "repo", "OPEN", 1)
	require.NoError(t, err)
	require.Len(t, prs, 2, "all pages must be accumulated")
	assert.Equal(t, 1, prs[0].ID)
	assert.Equal(t, 2, prs[1].ID)
	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount),
		"both pages must be fetched")
}

// TestCloudClient_ListRepos_MalformedNextDoesNotPanic verifies that when the
// "next" field contains a non-parseable URL the client still returns the
// accumulated values without panicking or returning an error.
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

	require.NotPanics(t, func() {
		repos, err := client.ListRepos(1)
		require.NoError(t, err)
		require.Len(t, repos, 1)
		assert.Equal(t, "only", repos[0].Slug)
	})
}

// TestCloudClient_BasicAuth_WhenEmailAndTokenSet verifies that when both an
// email (auth user) and a token are provided, Basic auth is used — the correct
// mode for Atlassian API tokens which require Basic email:token.
func TestCloudClient_BasicAuth_WhenEmailAndTokenSet(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "the-token", "alice@example.com")
	_, _ = client.GetCurrentUser()
	require.True(t, strings.HasPrefix(gotAuth, "Basic "),
		"expected Basic auth when both email+token set, got %q", gotAuth)
}

// TestCloudClient_BearerAuth_WhenTokenOnly verifies that when only a token is
// provided (no auth user / email), Bearer auth is used — the correct mode for
// OAuth2 / workspace access tokens.
func TestCloudClient_BearerAuth_WhenTokenOnly(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "the-token", "")
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

// TestCloudClient_ApprovePR_NoContentType verifies that ApprovePR (nil body)
// sends no Content-Type header — Bitbucket Cloud returns HTTP 400 when an
// empty POST includes Content-Type.
func TestCloudClient_ApprovePR_NoContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	_ = client.ApprovePR("myworkspace", "my-service", 42)
	assert.Empty(t, gotCT,
		"ApprovePR must not send Content-Type; Bitbucket Cloud rejects empty POSTs with Content-Type")
}
