package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api"
)

func newTLSClient(t *testing.T, handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "test-token"})
	return client, srv
}

func TestNewClient_BearerAuthHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	})
	var v struct{}
	_ = client.GetJSON("/test", &v)
	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestNewClient_BasicAuthHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Username: "alice", Password: "secret"})
	var v struct{}
	_ = client.GetJSON("/test", &v)
	assert.True(t, strings.HasPrefix(gotAuth, "Basic "), "expected Basic auth, got %q", gotAuth)
}

func TestNewClient_NoAuthWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{})
	var v struct{}
	_ = client.GetJSON("/test", &v)
	assert.Empty(t, gotAuth)
}

func TestHTTPError_Error_ContainsStatus(t *testing.T) {
	t.Parallel()
	err := &api.HTTPError{StatusCode: 404, Message: "not found"}
	assert.Contains(t, err.Error(), "404")
}

func TestHTTPError_Error_ContainsMessage(t *testing.T) {
	t.Parallel()
	err := &api.HTTPError{StatusCode: 401, Message: "Unauthorized"}
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGetJSON_200_Decodes(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"key":"value"}`)
	})
	var v map[string]string
	require.NoError(t, client.GetJSON("/test", &v))
	assert.Equal(t, "value", v["key"])
}

func TestGetJSON_401_ReturnsHTTPError(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Unauthorized"}]}`)
	})
	var v struct{}
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
}

func TestGetJSON_404_ReturnsHTTPError(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"not found"}]}`)
	})
	var v struct{}
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func TestGetJSON_NetworkError(t *testing.T) {
	t.Parallel()
	// Use a client that points at a closed server.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	srv.Close() // close before making request
	var v struct{}
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
}

func TestPostJSON_SendsBody(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":1}`)
	})
	var result map[string]int
	require.NoError(t, client.PostJSON("/repos", map[string]string{"name": "test"}, &result))
	assert.Contains(t, string(gotBody), "test")
}

func TestDelete_SendsDELETE(t *testing.T) {
	t.Parallel()
	var gotMethod string
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, client.Delete("/repos/PROJ/repo"))
	assert.Equal(t, http.MethodDelete, gotMethod)
}
