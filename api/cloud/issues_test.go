package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

const issueListJSON = `{
  "values": [
    {
      "id": 7,
      "title": "Crash on save",
      "state": "open",
      "kind": "bug",
      "priority": "major",
      "reporter": {"username":"alice","display_name":"Alice"},
      "assignee": {"username":"bob","display_name":"Bob"},
      "created_on": "2026-01-15T10:00:00.000000+00:00",
      "updated_on": "2026-01-16T11:00:00.000000+00:00",
      "links": {"html": {"href": "https://bitbucket.org/acme/repo/issues/7"}},
      "content": {"raw": "Crashes when ..."}
    },
    {
      "id": 8,
      "title": "Feature request",
      "state": "new",
      "kind": "enhancement",
      "priority": "minor",
      "reporter": {"username":"carol","display_name":"Carol"},
      "assignee": null,
      "created_on": "2026-02-01T08:00:00.000000+00:00",
      "updated_on": "2026-02-01T08:00:00.000000+00:00",
      "links": {"html": {"href": "https://bitbucket.org/acme/repo/issues/8"}},
      "content": {"raw": ""}
    }
  ]
}`

const issueGetJSON = `{
  "id": 42,
  "title": "Bug",
  "state": "open",
  "kind": "bug",
  "priority": "major",
  "reporter": {"username":"alice","display_name":"Alice"},
  "assignee": null,
  "created_on": "2026-01-15T10:00:00.000000+00:00",
  "updated_on": "2026-01-15T10:00:00.000000+00:00",
  "links": {"html": {"href": "https://bitbucket.org/acme/repo/issues/42"}},
  "content": {"raw": "body"}
}`

func newCloudIssueServer(t *testing.T, handler http.HandlerFunc) (*cloud.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", ""), srv
}

func TestCloudClient_ListIssues_PathAndQuery(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListIssues("acme", "repo", "open", 25)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/acme/repo/issues", gotPath)
	// Query order is alphabetical thanks to url.Values.Encode.
	assert.Contains(t, gotQuery, `pagelen=25`)
	assert.Contains(t, gotQuery, `q=state%3D%22open%22`)
	assert.Contains(t, gotQuery, `sort=-created_on`)
}

func TestCloudClient_ListIssues_OmitsStateFilterWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotQuery string
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListIssues("acme", "repo", "", 0)
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "q=state", "empty state must not produce a q= filter")
	assert.NotContains(t, gotQuery, "pagelen", "limit=0 must omit pagelen")
}

func TestCloudClient_ListIssues_Decodes(t *testing.T) {
	t.Parallel()
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(issueListJSON))
	})
	got, err := client.ListIssues("acme", "repo", "", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, 7, got[0].ID)
	assert.Equal(t, "Crash on save", got[0].Title)
	assert.Equal(t, "open", got[0].State)
	assert.Equal(t, "bug", got[0].Kind)
	assert.Equal(t, "alice", got[0].Reporter.Slug)
	require.NotNil(t, got[0].Assignee)
	assert.Equal(t, "bob", got[0].Assignee.Slug)
	assert.Equal(t, "https://bitbucket.org/acme/repo/issues/7", got[0].WebURL)

	// Unassigned: assignee:null must produce a nil pointer, not a zero User.
	assert.Nil(t, got[1].Assignee, "null assignee must decode to nil pointer")
}

func TestCloudClient_GetIssue_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(issueGetJSON))
	})
	got, err := client.GetIssue("acme", "repo", 42)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/acme/repo/issues/42", gotPath)
	assert.Equal(t, 42, got.ID)
	assert.Equal(t, "body", got.Content)
}

func TestCloudClient_CreateIssue_BodyAndPath(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody []byte
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(issueGetJSON))
	})

	got, err := client.CreateIssue("acme", "repo", backend.CreateIssueInput{
		Title:    "New bug",
		Content:  "Repro steps...",
		Kind:     "bug",
		Priority: "major",
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/repositories/acme/repo/issues", gotPath)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "New bug", sent["title"])
	assert.Equal(t, "bug", sent["kind"])
	assert.Equal(t, "major", sent["priority"])
	content, ok := sent["content"].(map[string]any)
	require.True(t, ok, "content must be a JSON object")
	assert.Equal(t, "Repro steps...", content["raw"])
	assert.Equal(t, 42, got.ID)
}

func TestCloudClient_CreateIssue_OmitsEmptyContent(t *testing.T) {
	// Bitbucket rejects content:{raw:""} on creation in some accounts;
	// omitempty + a pointer means an empty Content cleanly drops out.
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(issueGetJSON))
	})
	_, err := client.CreateIssue("acme", "repo", backend.CreateIssueInput{Title: "T"})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), "content")
	assert.NotContains(t, string(gotBody), "kind")
	assert.NotContains(t, string(gotBody), "priority")
}

func TestCloudClient_CreateIssue_RejectsEmptyTitle(t *testing.T) {
	t.Parallel()
	// No HTTP server: an empty title must short-circuit without dialing.
	client := cloud.NewClient(http.DefaultClient, "https://example.invalid", "tok", "")
	_, err := client.CreateIssue("acme", "repo", backend.CreateIssueInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title required")
}

func TestCloudClient_UpdateIssue_PathAndStateBody(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody []byte
	client, _ := newCloudIssueServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(issueGetJSON))
	})
	_, err := client.UpdateIssue("acme", "repo", 42, backend.UpdateIssueInput{State: "closed"})
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
	assert.Equal(t, "/repositories/acme/repo/issues/42", gotPath)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "closed", sent["state"])
	// Empty fields must not appear — the close path should send only state.
	_, hasTitle := sent["title"]
	assert.False(t, hasTitle, "empty fields must omitempty")
}
