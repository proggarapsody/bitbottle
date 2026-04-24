package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

// TestServerClient_ListRepos_FirstPageWithNextPageStart verifies that when a
// server response advertises more pages (isLastPage=false, nextPageStart=25)
// the current adapter returns the values from the first page without making
// additional requests. This documents the current single-page behaviour.
func TestServerClient_ListRepos_FirstPageWithNextPageStart(t *testing.T) {
	t.Parallel()
	var requestCount int32
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		body := `{"values":[
			{"id":1,"slug":"repo-a","name":"repo-a","project":{"key":"MYPROJ"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/repo-a/browse"}]}},
			{"id":2,"slug":"repo-b","name":"repo-b","project":{"key":"MYPROJ"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/repo-b/browse"}]}}
		],"size":2,"isLastPage":false,"start":0,"nextPageStart":25,"limit":25}`
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	repos, err := client.ListRepos(25)
	require.NoError(t, err)
	require.Len(t, repos, 2)
	assert.Equal(t, "repo-a", repos[0].Slug)
	assert.Equal(t, "repo-b", repos[1].Slug)
	// The initial request sends limit=25 (keyset not yet engaged).
	assert.Contains(t, gotQuery, "limit=25")
	assert.NotContains(t, gotQuery, "start=")
	// Single-page behaviour: no follow-up request is issued.
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

// TestServerClient_ListRepos_LimitQueryIncludedOnFirstRequest locks in that
// the first page request contains `limit=<N>` in the query string.
func TestServerClient_ListRepos_LimitQueryIncludedOnFirstRequest(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"size":0,"isLastPage":true,"start":0,"limit":50}`)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListRepos(50)
	require.NoError(t, err)
	assert.Equal(t, "limit=50", gotQuery)
}

// TestServerClient_ListPRs_FirstPageWithNextPageStart asserts PR listing
// also returns first-page values when more pages are advertised.
func TestServerClient_ListPRs_FirstPageWithNextPageStart(t *testing.T) {
	t.Parallel()
	var requestCount int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		body := `{"values":[
			{"id":1,"title":"PR1","state":"OPEN","author":{"user":{"slug":"alice","displayName":"Alice"}},"fromRef":{"id":"refs/heads/feat/a","displayId":"feat/a"},"toRef":{"id":"refs/heads/main","displayId":"main"},"links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/1"}]}}
		],"size":1,"isLastPage":false,"start":0,"nextPageStart":25,"limit":25}`
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, 1, prs[0].ID)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

// TestServerClient_BearerAuth_Precedence verifies that when both a token and
// a username are supplied Bearer wins.
func TestServerClient_BearerAuth_Precedence(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "the-token", "alice")
	_, _ = client.GetRepo("P", "r")
	assert.Equal(t, "Bearer the-token", gotAuth)
}

// TestServerClient_BasicAuth_WhenTokenEmpty verifies the Basic auth fallback
// when the token is empty but a username is supplied (password is empty).
func TestServerClient_BasicAuth_WhenTokenEmpty(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "", "alice")
	_, _ = client.GetRepo("P", "r")
	require.True(t, strings.HasPrefix(gotAuth, "Basic "),
		"expected Basic auth header, got %q", gotAuth)
}

// TestServerClient_NoAuth_WhenBothEmpty verifies no Authorization header when
// neither token nor username is supplied.
func TestServerClient_NoAuth_WhenBothEmpty(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "", "")
	_, _ = client.GetRepo("P", "r")
	assert.Empty(t, gotAuth)
}
