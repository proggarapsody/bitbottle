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

// TestGetJSON_MalformedResponseBody verifies that a 200 with invalid JSON returns
// a decode error rather than silently returning a zero value.
func TestGetJSON_MalformedResponseBody(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{not valid json`)
	})
	var v map[string]string
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
}

// TestHTTPError_EmptyErrorBody verifies that an empty error body (no JSON) still
// returns an HTTPError with the correct status code and an empty message.
func TestHTTPError_EmptyErrorBody(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	var v struct{}
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Empty(t, httpErr.Message)
}

// TestHTTPError_500_ReturnsHTTPError verifies that a 5xx response is treated as
// an HTTPError (not silently decoded into a zero-value struct).
func TestHTTPError_500_ReturnsHTTPError(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"errors":[{"message":"internal error"}]}`)
	})
	var v struct{}
	err := client.GetJSON("/test", &v)
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 500, httpErr.StatusCode)
	assert.Equal(t, "internal error", httpErr.Message)
}

// TestGetText_ReturnsRawBody verifies that GetText returns the raw response body
// as a string without JSON decoding.
func TestGetText_ReturnsRawBody(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "diff --git a/foo.go b/foo.go\n+added line\n")
	})
	text, err := client.GetText("/diff")
	require.NoError(t, err)
	assert.Contains(t, text, "diff --git")
	assert.Contains(t, text, "+added line")
}

// TestGetText_ErrorResponse verifies GetText surfaces HTTP errors correctly.
func TestGetText_ErrorResponse(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"diff not found"}]}`)
	})
	_, err := client.GetText("/diff")
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

// TestNewClient_MultipleTrailingSlashes verifies that even multiple trailing
// slashes are fully stripped by NewClient.
func TestNewClient_MultipleTrailingSlashes(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(srv.Close)

	// Three trailing slashes — all should be stripped.
	client := api.NewClient(srv.Client(), srv.URL+"///", api.AuthConfig{Token: "tok"})
	var v struct{}
	_ = client.GetJSON("/test", &v)
	assert.NotContains(t, gotPath, "//", "multiple trailing slashes should be stripped")
}

// TestPutJSON_SendsPUT verifies that PutJSON uses HTTP PUT method.
func TestPutJSON_SendsPUT(t *testing.T) {
	t.Parallel()
	var gotMethod string
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	var v struct{}
	require.NoError(t, client.PutJSON("/test", map[string]string{"k": "v"}, &v))
	assert.Equal(t, http.MethodPut, gotMethod)
}

// TestDelete_404_ReturnsHTTPError verifies that Delete on a missing resource
// returns HTTPError rather than nil.
func TestDelete_404_ReturnsHTTPError(t *testing.T) {
	t.Parallel()
	client, _ := newTLSClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"not found"}]}`)
	})
	err := client.Delete("/repos/PROJ/missing")
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}
