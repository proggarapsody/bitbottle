package server_test

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

	"github.com/proggarapsody/bitbottle/api/server"
)

// TestServerClient_ListRepos_FollowsNextPageStart verifies that when a server
// response advertises more pages (isLastPage=false, nextPageStart=N) the
// client follows the cursor and accumulates values from all pages.
func TestServerClient_ListRepos_FollowsNextPageStart(t *testing.T) {
	t.Parallel()
	var requestCount int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			// First page: two repos, more pages available.
			// The client should follow with start=25.
			body := `{"values":[
				{"id":1,"slug":"repo-a","name":"repo-a","project":{"key":"MYPROJ"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/repo-a/browse"}]}},
				{"id":2,"slug":"repo-b","name":"repo-b","project":{"key":"MYPROJ"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/repo-b/browse"}]}}
			],"size":2,"isLastPage":false,"start":0,"nextPageStart":25,"limit":25}`
			_, _ = io.WriteString(w, body)
		} else {
			// Second page: one more repo, no further pages.
			body := `{"values":[
				{"id":3,"slug":"repo-c","name":"repo-c","project":{"key":"MYPROJ"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/repo-c/browse"}]}}
			],"size":1,"isLastPage":true,"start":25,"nextPageStart":null,"limit":25}`
			_, _ = io.WriteString(w, body)
		}
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	repos, err := client.ListRepos(25)
	require.NoError(t, err)
	require.Len(t, repos, 3, "all pages must be accumulated")
	assert.Equal(t, "repo-a", repos[0].Slug)
	assert.Equal(t, "repo-b", repos[1].Slug)
	assert.Equal(t, "repo-c", repos[2].Slug)
	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount),
		"both pages must be fetched")
}

// TestServerClient_ListRepos_SecondPageHasStartParam verifies that the
// follow-up request carries start=<nextPageStart> in the query string.
func TestServerClient_ListRepos_SecondPageHasStartParam(t *testing.T) {
	t.Parallel()
	var secondQuery string
	var requestCount int32
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			body := fmt.Sprintf(`{"values":[
				{"id":1,"slug":"repo-a","name":"repo-a","project":{"key":"P"},"scmId":"git","state":"AVAILABLE","links":{"self":[{"href":"%s/browse"}]}}
			],"size":1,"isLastPage":false,"start":0,"nextPageStart":50,"limit":50}`, srv.URL)
			_, _ = io.WriteString(w, body)
		} else {
			secondQuery = r.URL.RawQuery
			_, _ = io.WriteString(w, `{"values":[],"size":0,"isLastPage":true,"start":50,"limit":50}`)
		}
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, _ = client.ListRepos(50)
	assert.Contains(t, secondQuery, "start=50",
		"second page request must include start=<nextPageStart>")
	assert.Contains(t, secondQuery, "limit=50",
		"second page request must preserve the original limit parameter")
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

// TestServerClient_ListPRs_FollowsNextPageStart asserts PR listing follows
// pagination across all available pages.
func TestServerClient_ListPRs_FollowsNextPageStart(t *testing.T) {
	t.Parallel()
	var requestCount int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			body := `{"values":[
				{"id":1,"title":"PR1","state":"OPEN","author":{"user":{"slug":"alice","displayName":"Alice"}},"fromRef":{"id":"refs/heads/feat/a","displayId":"feat/a"},"toRef":{"id":"refs/heads/main","displayId":"main"},"links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/1"}]}}
			],"size":1,"isLastPage":false,"start":0,"nextPageStart":25,"limit":25}`
			_, _ = io.WriteString(w, body)
		} else {
			body := `{"values":[
				{"id":2,"title":"PR2","state":"OPEN","author":{"user":{"slug":"bob","displayName":"Bob"}},"fromRef":{"id":"refs/heads/feat/b","displayId":"feat/b"},"toRef":{"id":"refs/heads/main","displayId":"main"},"links":{"self":[{"href":"https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/2"}]}}
			],"size":1,"isLastPage":true,"start":25,"limit":25}`
			_, _ = io.WriteString(w, body)
		}
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 2, "all pages must be accumulated")
	assert.Equal(t, 1, prs[0].ID)
	assert.Equal(t, 2, prs[1].ID)
	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount),
		"both pages must be fetched")
}

// TestServerClient_BasicAuth_WhenUsernameAndTokenSet verifies that when both a
// username and a token are provided Basic auth is used.
func TestServerClient_BasicAuth_WhenUsernameAndTokenSet(t *testing.T) {
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
	require.True(t, strings.HasPrefix(gotAuth, "Basic "),
		"expected Basic auth when both username+token set, got %q", gotAuth)
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
