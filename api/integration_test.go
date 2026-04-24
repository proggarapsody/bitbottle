package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api"
)

// TestClient_Integration_TrailingSlashBaseURL verifies the Step-7 refactor:
// a trailing slash in baseURL must not produce double-slash paths in real requests.
func TestClient_Integration_TrailingSlashBaseURL(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	// Intentionally pass a trailing slash — NewClient must strip it.
	client := api.NewClient(srv.Client(), srv.URL+"/", api.AuthConfig{Token: "tok"})
	_, err := client.ListRepos(10)
	require.NoError(t, err)
	assert.NotContains(t, gotPath, "//", "double slash in request path")
}

// TestClient_Integration_BearerTokenHeader verifies that a token in AuthConfig
// reaches the outgoing request as `Authorization: Bearer <token>`.
func TestClient_Integration_BearerTokenHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"alice","displayName":"Alice"}`)
	}))
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "my-secret-token"})
	_, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "Bearer my-secret-token", gotAuth)
}

// TestClient_Integration_BasicAuthHeader verifies username/password basic auth
// is sent when no token is present.
func TestClient_Integration_BasicAuthHeader(t *testing.T) {
	t.Parallel()
	var gotUser, gotPass string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"bob","displayName":"Bob"}`)
	}))
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Username: "bob", Password: "s3cr3t"})
	_, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "bob", gotUser)
	assert.Equal(t, "s3cr3t", gotPass)
}

// TestClient_Integration_HTTPErrorRequestURL verifies that HTTPError.RequestURL
// is populated from the actual request URL.
func TestClient_Integration_HTTPErrorRequestURL(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"repo not found"}]}`)
	}))
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	_, err := client.GetRepo("PROJ", "missing-repo")
	require.Error(t, err)

	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	assert.Contains(t, httpErr.RequestURL, srv.URL, "RequestURL should contain server URL")
	assert.Contains(t, httpErr.Message, "repo not found")
}

// TestClient_Integration_PostJSONContentType verifies that POST requests set
// Content-Type: application/json on the wire.
func TestClient_Integration_PostJSONContentType(t *testing.T) {
	t.Parallel()
	var gotContentType string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":1,"title":"new","state":"OPEN","author":{},"reviewers":[],"fromRef":{},"toRef":{},"links":{}}`)
	}))
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	_, err := client.CreatePR("PROJ", "repo", api.CreatePRInput{Title: "new"})
	require.NoError(t, err)
	assert.Equal(t, "application/json", gotContentType)
}

// TestClient_Integration_DeleteMethodOnWire verifies that DeleteBranch sends
// HTTP DELETE (not GET/POST), confirming the http.MethodDelete refactor.
func TestClient_Integration_DeleteMethodOnWire(t *testing.T) {
	t.Parallel()
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	err := client.DeleteBranch("PROJ", "repo", "refs/heads/feat/old")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
}
