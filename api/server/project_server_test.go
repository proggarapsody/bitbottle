package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func pagedProjects(projects ...map[string]any) map[string]any {
	values := make([]any, len(projects))
	for i, p := range projects {
		values[i] = p
	}
	return map[string]any{
		"values": values, "isLastPage": true, "size": len(projects),
	}
}

func fakeProject(key, name string) map[string]any {
	return map[string]any{
		"key":         key,
		"name":        name,
		"description": "Test project",
		"public":      false,
		"links": map[string]any{
			"self": []any{
				map[string]any{"href": "https://example.com/projects/" + key},
			},
		},
	}
}

// ── ListServerProjects ────────────────────────────────────────────────────────

func TestServerClient_ListServerProjects(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedProjects(fakeProject("PRJ", "My Project"), fakeProject("DEV", "Dev Project")))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	projects, err := c.ListServerProjects("", 25)
	require.NoError(t, err)
	assert.Contains(t, seenPath, "/projects")
	assert.Len(t, projects, 2)
	assert.Equal(t, "PRJ", projects[0].Key)
	assert.Equal(t, "DEV", projects[1].Key)
}

func TestServerClient_ListServerProjects_WithFilter(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pagedProjects(fakeProject("PRJ", "My Project")))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	projects, err := c.ListServerProjects("My", 25)
	require.NoError(t, err)
	assert.Contains(t, seenPath, "name=My")
	assert.Len(t, projects, 1)
}

// ── GetServerProject ──────────────────────────────────────────────────────────

func TestServerClient_GetServerProject(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeProject("PRJ", "My Project"))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	p, err := c.GetServerProject("PRJ")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PRJ", seenPath)
	assert.Equal(t, "PRJ", p.Key)
	assert.Equal(t, "My Project", p.Name)
	assert.Equal(t, "https://example.com/projects/PRJ", p.WebURL)
}

// ── CreateServerProject ───────────────────────────────────────────────────────

func TestServerClient_CreateServerProject(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fakeProject("NEWPRJ", "New Project"))
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	p, err := c.CreateServerProject(backend.CreateServerProjectInput{
		Key:         "NEWPRJ",
		Name:        "New Project",
		Description: "A new project",
		Public:      false,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Equal(t, "/projects", seenPath)
	assert.Equal(t, "NEWPRJ", seenBody["key"])
	assert.Equal(t, "New Project", seenBody["name"])
	assert.Equal(t, "NORMAL", seenBody["type"])
	assert.Equal(t, "NEWPRJ", p.Key)
}

// ── UpdateServerProject ───────────────────────────────────────────────────────

func TestServerClient_UpdateServerProject(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(fakeProject("PRJ", "Old Name"))
			return
		}
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		updated := fakeProject("PRJ", "New Name")
		_ = json.NewEncoder(w).Encode(updated)
	}))
	t.Cleanup(srv.Close)

	newName := "New Name"
	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	p, err := c.UpdateServerProject("PRJ", backend.UpdateServerProjectInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, seenMethod)
	assert.Equal(t, "/projects/PRJ", seenPath)
	assert.Equal(t, "New Name", seenBody["name"])
	assert.Equal(t, "PRJ", p.Key)
}

// ── DeleteServerProject ───────────────────────────────────────────────────────

func TestServerClient_DeleteServerProject(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	err := c.DeleteServerProject("PRJ")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Equal(t, "/projects/PRJ", seenPath)
}

// ── Interface assertion ───────────────────────────────────────────────────────

func TestServer_ServerProject_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.ServerProjectClient = (*server.Client)(nil)
}
