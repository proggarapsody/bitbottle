package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_CreatePR_SerialisesReviewers(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"title":"T","state":"OPEN","links":{"html":{"href":"https://bitbucket.org/x/y/pull-requests/1"}},"source":{"branch":{"name":"feat/x"}},"destination":{"branch":{"name":"main"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.CreatePR("acme", "my-service", backend.CreatePRInput{
		Title:      "T",
		FromBranch: "feat/x",
		ToBranch:   "main",
		Reviewers:  []string{"alice", "bob"},
	})
	require.NoError(t, err)

	var sent struct {
		Reviewers []struct {
			Username string `json:"username"`
		} `json:"reviewers"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	require.Len(t, sent.Reviewers, 2)
	assert.Equal(t, "alice", sent.Reviewers[0].Username)
	assert.Equal(t, "bob", sent.Reviewers[1].Username)
}

func TestCloudClient_CreatePR_NoReviewers_OmitsField(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"title":"T","state":"OPEN","links":{"html":{"href":"x"}},"source":{"branch":{"name":"x"}},"destination":{"branch":{"name":"main"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.CreatePR("acme", "my-service", backend.CreatePRInput{
		Title:      "T",
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), "reviewers")
}
