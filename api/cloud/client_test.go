package cloud_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newCloudClient(t *testing.T, handler http.HandlerFunc) (*cloud.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "test-token", "")
	return client, srv
}

func TestCloudClient_BearerAuth(t *testing.T) {
	t.Parallel()
	var gotAuth string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	})
	_, _ = client.GetCurrentUser()
	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestCloudClient_ErrorBody_ParsesTypeErrorFormat(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"type":"error","error":{"message":"Not Found"}}`)
	})
	_, err := client.GetCurrentUser()
	require.Error(t, err)
	var httpErr *backend.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
	assert.Equal(t, "Not Found", httpErr.Message)
}

func TestCloudClient_ErrorBody_MissingMessage(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"type":"error"}`)
	})
	_, err := client.GetCurrentUser()
	require.Error(t, err)
	var httpErr *backend.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 401, httpErr.StatusCode)
	// When the body has no parseable message, the transport falls back to the
	// canonical HTTP status text so the error is never just "HTTP 401: ".
	assert.Equal(t, "Unauthorized", httpErr.Message)
}
