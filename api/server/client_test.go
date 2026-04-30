package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// newServerClient creates a server.Client backed by the given TLS test server.
func newServerClient(t *testing.T, handler http.HandlerFunc) (*server.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "test-token", "")
	return client, srv
}

func TestServerClient_BearerAuth(t *testing.T) {
	t.Parallel()
	var gotAuth string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	})
	_, _ = client.GetRepo("P", "r") // triggers a request
	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestServerClient_BasicAuth(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)
	// Token empty, username set → Basic auth
	client := server.NewClient(srv.Client(), srv.URL, "", "alice")
	_, _ = client.GetRepo("P", "r")
	assert.True(t, strings.HasPrefix(gotAuth, "Basic "), "expected Basic auth header, got %q", gotAuth)
}

func TestServerClient_NoAuth(t *testing.T) {
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

func TestServerClient_HTTPError_Format(t *testing.T) {
	t.Parallel()
	err := &backend.HTTPError{StatusCode: 404, Message: "not found"}
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "not found")
}

func TestServerClient_GetJSON_200(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":1,"slug":"my-service","name":"my-service","project":{"key":"P"},"scmId":"git","links":{"self":[]}}`)
	})
	repo, err := client.GetRepo("P", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "my-service", repo.Slug)
}

func TestServerClient_GetJSON_401(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Unauthorized"}]}`)
	})
	_, err := client.GetRepo("P", "r")
	require.Error(t, err)
	var httpErr *backend.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

// TestServerClient_GetJSON_404_IsErrNotFound verifies that a 404 from the
// adapter surfaces as a typed backend.ErrNotFound, with the underlying
// HTTPError still reachable via errors.As. PRD #47, audit concern 5.
func TestServerClient_GetJSON_404_IsErrNotFound(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Repository not found"}]}`)
	})
	_, err := client.GetRepo("P", "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrNotFound)
	var httpErr *backend.HTTPError
	require.ErrorAs(t, err, &httpErr,
		"underlying HTTPError must remain reachable for callers that need status detail")
	assert.Equal(t, 404, httpErr.StatusCode)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.NotEmpty(t, de.Host, "host must be populated for MCP structured emission")
}

func TestServerClient_PostJSON_Body(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":2,"slug":"new-repo","name":"new-repo","project":{"key":"P"},"scmId":"git","links":{"self":[]}}`)
	})
	_, err := client.CreateRepo("P", backend.CreateRepoInput{Name: "new-repo", SCM: "git"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), "new-repo")
}

func TestServerClient_Delete_Method(t *testing.T) {
	t.Parallel()
	var gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteRepo("P", "repo")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
}
