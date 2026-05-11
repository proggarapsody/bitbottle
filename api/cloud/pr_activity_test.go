package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const cloudPRActivityJSON = `{"values":[
  {"approval":{"date":"2026-04-24T10:00:00+00:00","user":{"account_id":"123","display_name":"Alice","nickname":"alice"}}},
  {"comment":{"created_on":"2026-04-24T11:00:00Z","author":{"account_id":"456","display_name":"Bob","nickname":"bob"}}},
  {"update":{"date":"2026-04-24T12:00:00+00:00","author":{"account_id":"123","display_name":"Alice","nickname":"alice"}}}
]}`

func TestCloudClient_GetPRActivity(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudPRActivityJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	events, err := client.GetPRActivity("myws", "my-svc", 42, 0)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/activity", gotPath)
	require.Len(t, events, 3)

	assert.Equal(t, "approval", events[0].Type)
	assert.Equal(t, "alice", events[0].Actor.Slug)
	assert.Equal(t, "Alice", events[0].Actor.DisplayName)
	assert.False(t, events[0].CreatedAt.IsZero())

	assert.Equal(t, "comment", events[1].Type)
	assert.Equal(t, "bob", events[1].Actor.Slug)
	assert.False(t, events[1].CreatedAt.IsZero())

	assert.Equal(t, "update", events[2].Type)
	assert.Equal(t, "alice", events[2].Actor.Slug)
}

func TestCloudClient_GetPRActivity_SkipsUnknownEvents(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"some_future_field":{"x":1}},{"approval":{"date":"2026-04-24T10:00:00+00:00","user":{"account_id":"1","nickname":"alice"}}}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	events, err := client.GetPRActivity("ws", "repo", 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1, "unknown event type should be skipped")
	assert.Equal(t, "approval", events[0].Type)
}

func TestCloudClient_GetPRActivity_LimitCap(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudPRActivityJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	events, err := client.GetPRActivity("myws", "my-svc", 42, 2)
	require.NoError(t, err)
	assert.Len(t, events, 2, "limit cap should truncate results")
}

func TestCloudClient_GetPRActivity_DetailPopulated(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"approval":{"date":"2026-04-24T10:00:00+00:00","user":{"account_id":"1","nickname":"alice"}}}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	events, err := client.GetPRActivity("ws", "repo", 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.NotNil(t, events[0].Detail, "Detail should be populated")
	_, hasApproval := events[0].Detail["approval"]
	assert.True(t, hasApproval, "Detail should contain the approval sub-object")
}
