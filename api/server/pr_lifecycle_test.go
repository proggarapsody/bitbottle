package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_UpdatePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.UpdatePR("MYPROJ", "my-service", 42, backend.UpdatePRInput{Title: "New title"})
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42", gotPath)
}

func TestServerClient_UpdatePR_SendsBody(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.UpdatePR("MYPROJ", "my-service", 42, backend.UpdatePRInput{Title: "Updated", Description: "New body"})
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &payload))
	assert.Equal(t, "Updated", payload["title"])
	assert.Equal(t, "New body", payload["description"])
}

func TestServerClient_DeclinePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.DeclinePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/decline", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestServerClient_UnapprovePR_DeletesApproveEndpoint(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.UnapprovePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/approve", gotPath,
		"UnapprovePR must DELETE .../approve, not .../participants/~")
	assert.Equal(t, http.MethodDelete, gotMethod)
}

func TestServerClient_ReadyPR_GetsThenPutsFullBody(t *testing.T) {
	t.Parallel()
	var methods []string
	var bodies []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.ReadyPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	// First call GET, second call PUT with full body.
	require.Len(t, methods, 2)
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodPut, methods[1])
	put := bodies[1]
	assert.Contains(t, put, `"draft":false`)
	assert.Contains(t, put, `"title":"Fix login bug"`)
	assert.Contains(t, put, `"fromRef"`)
	assert.Contains(t, put, `"toRef"`)
}

func TestServerClient_RequestReview_GetsAndPutsPR(t *testing.T) {
	t.Parallel()
	var methods []string
	var bodies []string
	callCount := 0
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		callCount++
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	err := client.RequestReview("MYPROJ", "my-service", 42, []string{"alice", "bob"})
	require.NoError(t, err)
	// First call GET, second call PUT
	require.Len(t, methods, 2)
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodPut, methods[1])
	// PUT body should contain reviewers
	assert.Contains(t, bodies[1], "alice")
	assert.Contains(t, bodies[1], "bob")
}
