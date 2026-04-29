package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

func newTestTransport(t *testing.T, handler http.HandlerFunc) (*httpx.Transport, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	tr := httpx.New(srv.Client(), srv.URL, httpx.Auth{Token: "tok"}, nil)
	return tr, srv
}

// TestPostJSON_NilBody_SetsContentType verifies that a POST with no body still
// carries Content-Type: application/json so Bitbucket Server's CSRF protection
// does not reject the request with "XSRF check failed".
func TestPostJSON_NilBody_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	var result struct{}
	require.NoError(t, tr.PostJSON("/test", nil, &result))
	assert.Equal(t, "application/json", gotCT,
		"POST with nil body must still set Content-Type to pass Bitbucket Server CSRF check")
}

// TestDeleteJSON_NilBody_SetsContentType verifies the same CSRF fix for DELETE.
func TestDeleteJSON_NilBody_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, tr.DeleteJSON("/test", nil))
	assert.Equal(t, "application/json", gotCT,
		"DELETE with nil body must still set Content-Type to pass Bitbucket Server CSRF check")
}

// TestGetJSON_DoesNotSetContentType verifies GET requests are unaffected.
func TestGetJSON_DoesNotSetContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	var result struct{}
	require.NoError(t, tr.GetJSON("/test", &result))
	assert.Empty(t, gotCT, "GET must not set Content-Type")
}
