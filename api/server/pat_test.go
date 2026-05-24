package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func pagedPATs(pats ...map[string]any) map[string]any {
	values := make([]any, len(pats))
	for i, p := range pats {
		values[i] = p
	}
	return map[string]any{
		"values": values, "isLastPage": true, "size": len(pats),
	}
}

func fakePAT(id, name string) map[string]any {
	return map[string]any{
		"id":          id,
		"name":        name,
		"permissions": []any{"REPO_READ"},
		"createdDate": int64(1716556800000),
	}
}

// buildPATServer creates a TLS test server wiring the access-tokens REST root.
// The server dispatches on the path suffix after /rest/access-tokens/1.0.
func buildPATServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *server.Client) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	// NewClient wires patHTTP as schemeHost+"/rest/access-tokens/1.0";
	// we pass srv.URL as baseURL — the alt transport is schemeHost+suffix.
	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	return srv, c
}

// TestServerClient_ListPATs verifies list returns decoded PATs.
func TestServerClient_ListPATs(t *testing.T) {
	t.Parallel()
	var seenPath string
	_, c := buildPATServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedPATs(fakePAT("1", "CI Token"), fakePAT("2", "Dev Token")))
	})
	pats, err := c.ListPATs("alice", 25)
	require.NoError(t, err)
	assert.True(t, strings.Contains(seenPath, "/users/alice"), "path %q should contain /users/alice", seenPath)
	assert.Len(t, pats, 2)
	assert.Equal(t, "1", pats[0].ID)
	assert.Equal(t, "CI Token", pats[0].Name)
	assert.Equal(t, []string{"REPO_READ"}, pats[0].Permissions)
	// CreatedDate: 1716556800000 ms = 1716556800 s UTC
	assert.Equal(t, time.Unix(1716556800, 0).UTC(), pats[0].CreatedDate)
	assert.Nil(t, pats[0].ExpiryDate)
	assert.Nil(t, pats[0].LastUsed)
}

// TestServerClient_ListPATs_OptionalFields verifies expiryDate / lastAuthenticated mapping.
func TestServerClient_ListPATs_OptionalFields(t *testing.T) {
	t.Parallel()
	expiry := int64(1748092800000)
	lastAuth := int64(1716600000000)
	pat := map[string]any{
		"id":                "3",
		"name":              "With expiry",
		"permissions":       []any{"REPO_WRITE"},
		"createdDate":       int64(1716556800000),
		"expiryDate":        expiry,
		"lastAuthenticated": lastAuth,
	}
	_, c := buildPATServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedPATs(pat))
	})
	pats, err := c.ListPATs("alice", 10)
	require.NoError(t, err)
	require.Len(t, pats, 1)
	require.NotNil(t, pats[0].ExpiryDate)
	require.NotNil(t, pats[0].LastUsed)
	assert.Equal(t, time.Unix(1748092800, 0).UTC(), *pats[0].ExpiryDate)
	assert.Equal(t, time.Unix(1716600000, 0).UTC(), *pats[0].LastUsed)
}

// TestServerClient_CreatePAT verifies PUT + secret extraction.
func TestServerClient_CreatePAT(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	_, c := buildPATServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		resp := fakePAT("42", "CI Token")
		resp["token"] = "BBDC-supersecret"
		_ = json.NewEncoder(w).Encode(resp)
	})

	days := 30
	got, err := c.CreatePAT("alice", backend.CreatePATInput{
		Name:        "CI Token",
		Permissions: []string{"REPO_READ", "REPO_WRITE"},
		ExpiryDays:  &days,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, seenMethod)
	assert.True(t, strings.HasSuffix(seenPath, "/users/alice"), "path %q", seenPath)
	assert.Equal(t, "CI Token", seenBody["name"])
	assert.Equal(t, float64(30), seenBody["expiryDays"])
	assert.Equal(t, "42", got.ID)
	assert.Equal(t, "CI Token", got.Name)
	assert.Equal(t, "BBDC-supersecret", got.Token)
}

// TestServerClient_CreatePAT_NoExpiry verifies omitempty on expiryDays.
func TestServerClient_CreatePAT_NoExpiry(t *testing.T) {
	t.Parallel()
	var seenBody map[string]any
	_, c := buildPATServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		resp := fakePAT("5", "No Expiry")
		resp["token"] = "BBDC-abc"
		_ = json.NewEncoder(w).Encode(resp)
	})
	_, err := c.CreatePAT("alice", backend.CreatePATInput{
		Name:        "No Expiry",
		Permissions: []string{"REPO_READ"},
	})
	require.NoError(t, err)
	_, hasExpiry := seenBody["expiryDays"]
	assert.False(t, hasExpiry, "expiryDays should be absent when nil")
}

// TestServerClient_RevokePAT verifies DELETE path.
func TestServerClient_RevokePAT(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	_, c := buildPATServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.RevokePAT("alice", "42")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.True(t, strings.HasSuffix(seenPath, "/users/alice/42"), "path %q", seenPath)
}

// TestServer_PAT_ImplementsInterface is a compile-time assertion.
func TestServer_PAT_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.PATClient = (*server.Client)(nil)
}
