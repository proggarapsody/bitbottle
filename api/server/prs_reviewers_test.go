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

// TestServerClient_CreatePR_SerialisesReviewers proves the server adapter
// translates a flat []string of slugs into BBS's nested reviewers array
// shape: [{user: {name: SLUG}}].
func TestServerClient_CreatePR_SerialisesReviewers(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"version":0,"title":"T","state":"OPEN"}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "T",
		FromBranch: "feat/x",
		ToBranch:   "main",
		Reviewers:  []string{"alice", "bob"},
	})
	require.NoError(t, err)

	var sent struct {
		Reviewers []struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
		} `json:"reviewers"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	require.Len(t, sent.Reviewers, 2)
	assert.Equal(t, "alice", sent.Reviewers[0].User.Name)
	assert.Equal(t, "bob", sent.Reviewers[1].User.Name)
}

func TestServerClient_CreatePR_NoReviewers_OmitsField(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"version":0,"title":"T","state":"OPEN"}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "T",
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), "reviewers",
		"empty reviewers must omit the field entirely (omitempty)")
}
