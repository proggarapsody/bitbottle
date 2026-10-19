package server_test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_ListPRs_MapsAuthorSlug(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "alice", prs[0].Author.Slug)
}

func TestServerClient_ListPRs_MapsFromBranch(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "fix/login", prs[0].FromBranch)
}

func TestServerClient_ListPRs_RespectsTotalLimit(t *testing.T) {
	t.Parallel()
	page1 := `{"size":2,"isLastPage":false,"nextPageStart":2,"start":0,"values":[` +
		`{"id":1,"title":"a","author":{"user":{"slug":"u"}},"fromRef":{"id":"refs/heads/x"},"toRef":{"id":"refs/heads/main"},"links":{"self":[{"href":"http://h/1"}]}},` +
		`{"id":2,"title":"b","author":{"user":{"slug":"u"}},"fromRef":{"id":"refs/heads/x"},"toRef":{"id":"refs/heads/main"},"links":{"self":[{"href":"http://h/2"}]}}]}`
	page2 := `{"size":2,"isLastPage":true,"start":2,"values":[` +
		`{"id":3,"title":"c","author":{"user":{"slug":"u"}},"fromRef":{"id":"refs/heads/x"},"toRef":{"id":"refs/heads/main"},"links":{"self":[{"href":"http://h/3"}]}},` +
		`{"id":4,"title":"d","author":{"user":{"slug":"u"}},"fromRef":{"id":"refs/heads/x"},"toRef":{"id":"refs/heads/main"},"links":{"self":[{"href":"http://h/4"}]}}]}`
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("start") == "2" {
			_, _ = io.WriteString(w, page2)
			return
		}
		_, _ = io.WriteString(w, page1)
	})

	prs, err := client.ListPRs("MYPROJ", "my-service", "ALL", 2)
	require.NoError(t, err)
	assert.Len(t, prs, 2)
}

func TestServerClient_GetPR_MapsAllFields(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_get.json", 200)
	pr, err := client.GetPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, 42, pr.ID)
	assert.Equal(t, "Fix login bug", pr.Title)
	assert.Equal(t, "Fixes auth", pr.Description)
	assert.Equal(t, "OPEN", pr.State)
	assert.Equal(t, "alice", pr.Author.Slug)
	assert.Equal(t, "Alice", pr.Author.DisplayName)
	assert.Equal(t, "fix/login", pr.FromBranch)
	assert.Equal(t, "main", pr.ToBranch)
	assert.Contains(t, pr.WebURL, "pull-requests/42")
}

// TestServerClient_GetPR_CopiesVersion verifies that the version field from the
// Bitbucket Server wire response is forwarded into the domain PullRequest.
// pr_get.json contains "version":3.
func TestServerClient_GetPR_CopiesVersion(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_get.json", 200)
	pr, err := client.GetPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, 3, pr.Version, "toPRDomain must copy the version field from the wire response")
}

func TestServerClient_CreatePR_AddsRefsHeadsPrefix(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "Test PR",
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"id":"refs/heads/feat/x"`)
}

func TestServerClient_CreatePR_NoDoublePrefix(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "Test PR",
		FromBranch: "refs/heads/feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	bodyStr := string(gotBody)
	// Should appear exactly once, not refs/heads/refs/heads/
	assert.Contains(t, bodyStr, `"id":"refs/heads/feat/x"`)
	assert.NotContains(t, bodyStr, "refs/heads/refs/heads/")
}

// newMergePRHandler returns a handler that serves testdata/pr_get.json for all
// requests and captures the body of the final POST (the actual merge call).
func newMergePRHandler(t *testing.T, mergeBody *[]byte) http.HandlerFunc {
	t.Helper()
	fixture, err := os.ReadFile("testdata/pr_get.json")
	require.NoError(t, err)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			*mergeBody, _ = io.ReadAll(r.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}
}

func TestServerClient_MergePR_SendsStrategy(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, newMergePRHandler(t, &gotBody))
	_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{Strategy: "squash"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"strategyId":"squash"`)
	assert.NotContains(t, string(gotBody), `"strategy":`)
}

func TestServerClient_MergePR_StrategyIDs(t *testing.T) {
	for _, tc := range []struct{ input, expected string }{
		{"merge", "no-ff"},
		{"merge-commit", "no-ff"},
		{"squash", "squash"},
		{"rebase", "fast-forward"},
		{"custom-id", "custom-id"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var body []byte
			client, _ := newServerClient(t, newMergePRHandler(t, &body))
			_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{Strategy: tc.input})
			require.NoError(t, err)
			assert.Contains(t, string(body), `"strategyId":"`+tc.expected+`"`)
		})
	}
}

func TestServerClient_EnableAutoMerge_StrategyIDs(t *testing.T) {
	for _, tc := range []struct{ input, expected string }{
		{"merge", "no-ff"},
		{"merge-commit", "no-ff"},
		{"squash", "squash"},
		{"rebase", "fast-forward"},
		{"custom-id", "custom-id"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var body []byte
			client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					_, _ = io.WriteString(w, `{"version":3}`)
					return
				}
				body, _ = io.ReadAll(r.Body)
				_, _ = io.WriteString(w, `{}`)
			})
			err := client.EnableAutoMerge("MYPROJ", "my-service", 42, tc.input)
			require.NoError(t, err)
			assert.JSONEq(t, `{"version":3,"autoMerge":true,"strategyId":"`+tc.expected+`"}`, string(body))
		})
	}
}

func TestServerClient_GetPR_AutoMergeStrategyVocabulary(t *testing.T) {
	for _, tc := range []struct{ wire, expected string }{
		{"no-ff", "merge"},
		{"squash", "squash"},
		{"fast-forward", "rebase"},
		{"custom-id", "custom-id"},
	} {
		t.Run(tc.wire, func(t *testing.T) {
			client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"id":42,"autoMerge":{"strategyId":"`+tc.wire+`"}}`)
			})
			pr, err := client.GetPR("MYPROJ", "my-service", 42)
			require.NoError(t, err)
			require.NotNil(t, pr.AutoMerge)
			assert.Equal(t, tc.expected, pr.AutoMerge.Strategy)
		})
	}
}

func TestServerClient_MergePR_EmptyStrategyOmitted(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, newMergePRHandler(t, &gotBody))
	_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), `"strategyId"`)
}

// TestServerClient_MergePR_SendsCurrentVersion verifies that MergePR fetches
// the PR first and forwards its version in the merge POST body — the field
// Bitbucket Server requires to avoid HTTP 409 "out-of-date information".
func TestServerClient_MergePR_SendsCurrentVersion(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, newMergePRHandler(t, &gotBody))
	_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{})
	require.NoError(t, err)
	// pr_get.json has "version":3; the merge body must echo it back.
	assert.Contains(t, string(gotBody), `"version":3`)
}

func TestServerClient_ApprovePR_PostsToApproveEndpoint(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.ApprovePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod, "ApprovePR must POST, not PUT")
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/approve", gotPath)
}

func TestServerClient_DeleteBranch_SendsDeleteWithBody(t *testing.T) {
	t.Parallel()
	var gotMethod string
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteBranch("MYPROJ", "my-service", "feat/old")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Contains(t, string(gotBody), `"name"`)
}

// Queueing auto-merge reads the current version and posts to the merge endpoint.
func TestServerClient_EnableAutoMerge_SendsVersionedMerge(t *testing.T) {
	t.Parallel()
	var requests []string
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"version":3}`)
			return
		}
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.EnableAutoMerge("MYPROJ", "my-service", 42, "squash")
	require.NoError(t, err)
	assert.Equal(t, []string{"GET /projects/MYPROJ/repos/my-service/pull-requests/42", "POST /projects/MYPROJ/repos/my-service/pull-requests/42/merge"}, requests)
	assert.JSONEq(t, `{"version":3,"autoMerge":true,"strategyId":"squash"}`, string(gotBody))
}

func TestServerClient_GetCurrentUser_MapsSlug(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"alice","displayName":"Alice"}`)
	})
	user, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Slug)
}
