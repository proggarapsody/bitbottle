package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

const serverPRActivityJSON = `{"values":[
  {"action":"APPROVED","createdDate":1714000000000,"user":{"slug":"alice","displayName":"Alice"}},
  {"action":"COMMENTED","createdDate":1714000100000,"user":{"slug":"bob","displayName":"Bob"}},
  {"action":"UPDATED","createdDate":1714000200000,"user":{"slug":"alice","displayName":"Alice"}},
  {"action":"MERGED","createdDate":1714000300000,"user":{"slug":"alice","displayName":"Alice"}},
  {"action":"DECLINED","createdDate":1714000400000,"user":{"slug":"carol","displayName":"Carol"}},
  {"action":"RESCOPED","createdDate":1714000500000,"user":{"slug":"alice","displayName":"Alice"}},
  {"action":"UNAPPROVED","createdDate":1714000600000,"user":{"slug":"bob","displayName":"Bob"}}
],"isLastPage":true}`

func TestServerClient_GetPRActivity(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverPRActivityJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	events, err := client.GetPRActivity("MYPROJ", "my-svc", 42, 0)
	require.NoError(t, err)

	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/activities", gotPath)
	require.Len(t, events, 7)

	assert.Equal(t, "approval", events[0].Type)
	assert.Equal(t, "alice", events[0].Actor.Slug)
	assert.Equal(t, "Alice", events[0].Actor.DisplayName)
	assert.False(t, events[0].CreatedAt.IsZero())

	assert.Equal(t, "comment", events[1].Type)
	assert.Equal(t, "bob", events[1].Actor.Slug)

	assert.Equal(t, "update", events[2].Type)
	assert.Equal(t, "merge", events[3].Type)
	assert.Equal(t, "declined", events[4].Type)
	assert.Equal(t, "rescoped", events[5].Type)
	assert.Equal(t, "unapproval", events[6].Type)
}

func TestServerClient_GetPRActivity_SkipsUnknownActions(t *testing.T) {
	t.Parallel()
	const body = `{"values":[
    {"action":"OPENED","createdDate":1714000000000,"user":{"slug":"alice"}},
    {"action":"APPROVED","createdDate":1714000100000,"user":{"slug":"alice","displayName":"Alice"}}
  ],"isLastPage":true}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	events, err := client.GetPRActivity("PROJ", "repo", 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1, "unknown action should be skipped")
	assert.Equal(t, "approval", events[0].Type)
}

func TestServerClient_GetPRActivity_LimitCap(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverPRActivityJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	events, err := client.GetPRActivity("MYPROJ", "my-svc", 42, 3)
	require.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestServerClient_GetPRActivity_CreatedAtEpochMs(t *testing.T) {
	t.Parallel()
	const body = `{"values":[
    {"action":"APPROVED","createdDate":1714000000000,"user":{"slug":"alice"}}
  ],"isLastPage":true}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	events, err := client.GetPRActivity("PROJ", "repo", 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	// 1714000000000 ms = 1714000000 s
	assert.Equal(t, int64(1714000000), events[0].CreatedAt.Unix())
}

func TestServerClient_GetPRActivity_DetailPopulated(t *testing.T) {
	t.Parallel()
	const body = `{"values":[
    {"action":"APPROVED","createdDate":1714000000000,"user":{"slug":"alice"},"someExtraField":"x"}
  ],"isLastPage":true}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	events, err := client.GetPRActivity("PROJ", "repo", 1, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.NotNil(t, events[0].Detail)
	assert.Equal(t, "APPROVED", events[0].Detail["action"])
}
