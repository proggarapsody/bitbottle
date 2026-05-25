package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

const wsPageJSON = `{
  "values": [
    {"uuid":"{aaaaaaaa-1111-1111-1111-111111111111}","slug":"acme","name":"Acme Inc",
     "links":{"html":{"href":"https://bitbucket.org/acme/"}}},
    {"uuid":"{bbbbbbbb-2222-2222-2222-222222222222}","slug":"beta","name":"Beta Co",
     "links":{"html":{"href":"https://bitbucket.org/beta/"}}}
  ]
}`

const projectPageJSON = `{
  "values": [
    {"uuid":"{cccccccc-3333-3333-3333-333333333333}","key":"PROJ","name":"Project Alpha",
     "links":{"html":{"href":"https://bitbucket.org/acme/workspace/projects/PROJ"}}},
    {"uuid":"{dddddddd-4444-4444-4444-444444444444}","key":"BETA","name":"Project Beta",
     "links":{"html":{"href":"https://bitbucket.org/acme/workspace/projects/BETA"}}}
  ]
}`

func TestCloudClient_ListWorkspaces_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWorkspaces(0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces", gotPath)
}

func TestCloudClient_ListWorkspaces_IncludesPagelenWhenLimited(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWorkspaces(50)
	require.NoError(t, err)
	assert.Equal(t, "pagelen=50", gotQuery)
}

func TestCloudClient_ListWorkspaces_DecodesValues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(wsPageJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListWorkspaces(0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "acme", got[0].Slug)
	assert.Equal(t, "Acme Inc", got[0].Name)
	// UUID braces are stripped at the domain boundary so consumers don't have
	// to second-guess the format.
	assert.Equal(t, "aaaaaaaa-1111-1111-1111-111111111111", got[0].UUID)
	assert.Equal(t, "https://bitbucket.org/acme/", got[0].WebURL)
	assert.Equal(t, "beta", got[1].Slug)
}

func TestCloudClient_ListWorkspaces_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.ListWorkspaces(0)
	require.Error(t, err)
}

func TestCloudClient_ListProjects_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListProjects("acme", 0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/acme/projects", gotPath)
}

func TestCloudClient_SearchWorkspaces_NoFilters(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{})
	require.NoError(t, err)
	assert.Equal(t, "", gotQuery)
}

func TestCloudClient_SearchWorkspaces_WithQuery(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{Query: "myws"})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "myws")
	assert.Contains(t, gotQuery, "slug")
}

func TestCloudClient_SearchWorkspaces_WithRole(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{Role: "owner"})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "role=owner")
}

func TestCloudClient_SearchWorkspaces_WithLimit(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "pagelen=10")
}

func TestCloudClient_SearchWorkspaces_DecodesValues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(wsPageJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "acme", got[0].Slug)
	assert.Equal(t, "beta", got[1].Slug)
}

func TestCloudClient_SearchWorkspaces_LimitCappedAt50(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{Limit: 200})
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "pagelen=50")
}

func TestCloudClient_SearchWorkspaces_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.SearchWorkspaces(backend.WorkspaceSearchOpts{})
	require.Error(t, err)
}

func TestCloudClient_ListProjects_RejectsEmptyWorkspace(t *testing.T) {
	t.Parallel()
	// No HTTP server: an empty workspace must short-circuit before the call.
	client := cloud.NewClient(http.DefaultClient, "https://example.invalid", "tok", "")
	_, err := client.ListProjects("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestCloudClient_ListProjects_DecodesValues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(projectPageJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListProjects("acme", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "PROJ", got[0].Key)
	assert.Equal(t, "Project Alpha", got[0].Name)
	assert.Equal(t, "cccccccc-3333-3333-3333-333333333333", got[0].UUID)
	assert.Equal(t, "https://bitbucket.org/acme/workspace/projects/PROJ", got[0].WebURL)
	assert.Equal(t, "BETA", got[1].Key)
}
