package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

func buildServerPagedPRs(prs []map[string]any) []byte {
	body, _ := json.Marshal(map[string]any{
		"values":     prs,
		"isLastPage": true,
		"size":       len(prs),
		"start":      0,
	})
	return body
}

func makeServerPR(id int, title string) map[string]any {
	return map[string]any{
		"id":          id,
		"version":     1,
		"title":       title,
		"description": "",
		"state":       "OPEN",
		"draft":       false,
		"author": map[string]any{
			"user": map[string]any{
				"slug":        "alice",
				"displayName": "Alice",
			},
			"role": "AUTHOR",
		},
		"fromRef": map[string]any{
			"id":           "refs/heads/feat/x",
			"displayId":    "feat/x",
			"latestCommit": "abc1234",
		},
		"toRef": map[string]any{
			"id":        "refs/heads/main",
			"displayId": "main",
		},
		"links": map[string]any{
			"self": []map[string]any{
				{"href": fmt.Sprintf("https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/%d", id)},
			},
		},
	}
}

func makeServerInboxPR(id int, title, key, slug string) map[string]any {
	return map[string]any{
		"id":    id,
		"title": title,
		"state": "OPEN",
		"toRef": map[string]any{
			"repository": map[string]any{
				"project": map[string]any{"key": key},
				"slug":    slug,
			},
		},
		"fromRef": map[string]any{
			"latestCommit": "def5678",
		},
		"author": map[string]any{
			"user": map[string]any{
				"slug":        "bob",
				"displayName": "Bob",
			},
		},
		"links": map[string]any{
			"self": []map[string]any{
				{"href": fmt.Sprintf("https://bb.example.com/projects/%s/repos/%s/pull-requests/%d", key, slug, id)},
			},
		},
	}
}

func TestServerClient_ListMyPRs_ReturnsReviewerPRsFromInbox(t *testing.T) {
	t.Parallel()

	inboxPR := makeServerInboxPR(10, "Inbox PR", "PROJ", "repo")
	authorPR := makeServerPR(20, "Author PR")

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/inbox/pull-requests":
			_, _ = w.Write(buildServerPagedPRs([]map[string]any{inboxPR}))
		case "/users/~":
			_, _ = w.Write([]byte(`{"slug":"alice","displayName":"Alice"}`))
		default:
			_, _ = w.Write(buildServerPagedPRs([]map[string]any{authorPR}))
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	entries, err := client.ListMyPRs("PROJ", "repo")
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	var reviewerEntry, authorEntry *struct {
		ID   int
		Role string
	}
	for i := range entries {
		e := &struct {
			ID   int
			Role string
		}{ID: entries[i].ID, Role: entries[i].Role}
		if entries[i].Role == "REVIEWER" {
			reviewerEntry = e
		}
		if entries[i].Role == "AUTHOR" {
			authorEntry = e
		}
	}
	require.NotNil(t, reviewerEntry, "expected a REVIEWER entry")
	assert.Equal(t, 10, reviewerEntry.ID)
	require.NotNil(t, authorEntry, "expected an AUTHOR entry")
	assert.Equal(t, 20, authorEntry.ID)
}

func TestServerClient_ListMyPRs_AuthorWinsOnConflict(t *testing.T) {
	t.Parallel()

	// Same PR ID in inbox (REVIEWER) and author list (AUTHOR)
	inboxPR := makeServerInboxPR(42, "Reviewer version", "PROJ", "repo")
	authorPR := makeServerPR(42, "Author version")

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/inbox/pull-requests":
			_, _ = w.Write(buildServerPagedPRs([]map[string]any{inboxPR}))
		case "/users/~":
			_, _ = w.Write([]byte(`{"slug":"alice","displayName":"Alice"}`))
		default:
			_, _ = w.Write(buildServerPagedPRs([]map[string]any{authorPR}))
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	entries, err := client.ListMyPRs("PROJ", "repo")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "AUTHOR", entries[0].Role)
	assert.Equal(t, "Author version", entries[0].Title)
}
