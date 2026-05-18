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

func TestCloudClient_UnreadyPR_SendsDraftTrue(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(prGetBody(t))
	})
	err := client.UnreadyPR("myworkspace", "my-service", 7)
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"draft":true`)
}

// TestCloudClient_RequestReview_UsesPutWithReviewersList verifies that
// RequestReview issues exactly 2 requests: GET to read the current PR, then
// PUT to /pullrequests/{id} carrying the merged reviewers list.
func TestCloudClient_RequestReview_UsesPutWithReviewersList(t *testing.T) {
	t.Parallel()
	var methods []string
	var paths []string
	var putBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			putBody, _ = io.ReadAll(r.Body)
		}
		_, _ = w.Write(prGetBody(t))
	})
	err := client.RequestReview("myworkspace", "my-service", 7, []string{"alice", "bob"})
	require.NoError(t, err)
	require.Len(t, methods, 2, "expect exactly GET then PUT")
	assert.Equal(t, http.MethodGet, methods[0], "first call must be GET (read current PR)")
	assert.Equal(t, http.MethodPut, methods[1], "second call must be PUT (update reviewers)")
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/7", paths[1])
	assert.Contains(t, string(putBody), "alice")
	assert.Contains(t, string(putBody), "bob")
}

// TestCloudClient_RequestReview_PreservesExistingReviewers verifies that
// reviewers already on the PR are preserved when adding new ones.
func TestCloudClient_RequestReview_PreservesExistingReviewers(t *testing.T) {
	t.Parallel()
	var putBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			putBody, _ = io.ReadAll(r.Body)
		}
		// PR with an existing reviewer "charlie"
		_, _ = io.WriteString(w, `{"id":7,"title":"My PR","reviewers":[{"account_id":"charlie"}],"source":{"branch":{"name":"feat"}},"destination":{"branch":{"name":"main"}},"links":{"html":{"href":""}}}`)
	})
	err := client.RequestReview("myworkspace", "my-service", 7, []string{"alice"})
	require.NoError(t, err)
	assert.Contains(t, string(putBody), "charlie", "existing reviewer must be preserved")
	assert.Contains(t, string(putBody), "alice", "new reviewer must be added")
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
