package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func prGetBody(t *testing.T) []byte {
	t.Helper()
	return []byte(`{"id":7,"title":"My PR","description":"desc","state":"OPEN","draft":false,"author":{"display_name":"Alice","account_id":"alice"},"source":{"branch":{"name":"feat/x"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":"https://bitbucket.org/ws/repo/pull-requests/7"}}}`)
}

func TestCloudClient_UpdatePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(prGetBody(t))
	})
	_, err := client.UpdatePR("myworkspace", "my-service", 7, backend.UpdatePRInput{Title: "New title"})
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7", gotPath)
}

func TestCloudClient_UpdatePR_SendsBody(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(prGetBody(t))
	})
	_, err := client.UpdatePR("myworkspace", "my-service", 7, backend.UpdatePRInput{Title: "Updated", Description: "New body"})
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &payload))
	assert.Equal(t, "Updated", payload["title"])
	assert.Equal(t, "New body", payload["description"])
}

func TestCloudClient_DeclinePR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.DeclinePR("myworkspace", "my-service", 7)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7/decline", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestCloudClient_UnapprovePR_IssuesDeleteMethod(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.UnapprovePR("myworkspace", "my-service", 7)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7/approve", gotPath)
	assert.Equal(t, http.MethodDelete, gotMethod)
}

func TestCloudClient_ReadyPR_SendsDraftFalse(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(prGetBody(t))
	})
	err := client.ReadyPR("myworkspace", "my-service", 7)
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"draft":false`)
}

func TestCloudClient_RequestReview_SendsOneRequestPerUser(t *testing.T) {
	t.Parallel()
	var paths []string
	var bodies []string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.RequestReview("myworkspace", "my-service", 7, []string{"alice", "bob"})
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7/participants", paths[0])
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7/participants", paths[1])
	assert.Contains(t, bodies[0], "alice")
	assert.Contains(t, bodies[1], "bob")
}

func TestCloudClient_RequestChangesPR_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.RequestChangesPR("myworkspace", "my-service", 7)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7/request-changes", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}
