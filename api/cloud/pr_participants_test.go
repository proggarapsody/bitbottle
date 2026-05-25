package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listPRParticipantsJSON = `{"values":[
{"user":{"account_id":"u1","display_name":"Alice","nickname":"alice"},"role":"REVIEWER","approved":true,"state":"approved"},
{"user":{"account_id":"u2","display_name":"Bob","nickname":"bob"},"role":"REVIEWER","approved":false,"state":"changes_requested"},
{"user":{"account_id":"u3","display_name":"Carol","nickname":"carol"},"role":"PARTICIPANT","approved":false,"state":""}
]}`

const updatePRParticipantJSON = `{"user":{"account_id":"u1","display_name":"Alice","nickname":"alice"},"role":"REVIEWER","approved":true,"state":"approved"}`

func TestCloudClient_UpdatePRParticipant_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(updatePRParticipantJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.UpdatePRParticipant("myworkspace", "my-service", 42, "u1", "approved")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/42/participants/u1", gotPath)
}

func TestCloudClient_UpdatePRParticipant_MapsResponse(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(updatePRParticipantJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	p, err := client.UpdatePRParticipant("myworkspace", "my-service", 42, "u1", "approved")
	require.NoError(t, err)

	assert.Equal(t, "alice", p.User.Slug)
	assert.Equal(t, "Alice", p.User.DisplayName)
	assert.Equal(t, "REVIEWER", p.Role)
	assert.True(t, p.Approved)
	assert.Equal(t, "APPROVED", p.State)
}

func TestCloudClient_ListPRParticipants_IssuesCorrectPathAndNormalizesState(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listPRParticipantsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	parts, err := client.ListPRParticipants("myworkspace", "my-service", 42)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/42/participants?pagelen=100", gotPath)
	require.Len(t, parts, 3)

	assert.Equal(t, "REVIEWER", parts[0].Role)
	assert.True(t, parts[0].Approved)
	assert.Equal(t, "APPROVED", parts[0].State)

	assert.Equal(t, "REVIEWER", parts[1].Role)
	assert.False(t, parts[1].Approved)
	assert.Equal(t, "CHANGES_REQUESTED", parts[1].State)

	assert.Equal(t, "PARTICIPANT", parts[2].Role)
	assert.False(t, parts[2].Approved)
	assert.Equal(t, "", parts[2].State)
}
