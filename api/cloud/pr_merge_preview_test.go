package cloud_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudClient_DryRunMergePR_CanMerge(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"can_merge_without_conflicts":true,"message":"Looks good"}`)
	})
	result, err := client.DryRunMergePR("myws", "my-repo", 7, "squash")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-repo/pullrequests/7/merge?dry_run=true", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, string(gotBody), `"merge_strategy"`)
	assert.True(t, result.CanMerge)
	assert.Equal(t, "Looks good", result.Message)
}

func TestCloudClient_DryRunMergePR_NoStrategy_EmptyBody(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"can_merge_without_conflicts":true}`)
	})
	result, err := client.DryRunMergePR("myws", "my-repo", 7, "")
	require.NoError(t, err)
	assert.Empty(t, gotBody) // nil body → no strategy key
	assert.True(t, result.CanMerge)
}

func TestCloudClient_DryRunMergePR_ConflictReturns409(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"type":"error","error":{"message":"This pull request has conflicts"}}`)
	})
	result, err := client.DryRunMergePR("myws", "my-repo", 7, "")
	require.NoError(t, err)
	assert.False(t, result.CanMerge)
	assert.Contains(t, result.Message, "conflict")
}

func TestCloudClient_DryRunMergePR_PathEscaping(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"can_merge_without_conflicts":true}`)
	})
	_, err := client.DryRunMergePR("myws", "my-repo", 42, "")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-repo/pullrequests/42/merge", gotPath)
}
