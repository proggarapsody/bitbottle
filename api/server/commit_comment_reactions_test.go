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

// ── ListCommitCommentReactions ────────────────────────────────────────────────

func TestServerClient_ListCommitCommentReactions_GroupsByEmoji(t *testing.T) {
	t.Parallel()
	const body = `{
		"values": [
			{"emoticon": {"value": ":thumbsup:"}, "user": {"slug": "alice", "displayName": "Alice"}},
			{"emoticon": {"value": ":thumbsup:"}, "user": {"slug": "bob",   "displayName": "Bob"}},
			{"emoticon": {"value": ":heart:"},    "user": {"slug": "carol", "displayName": "Carol"}}
		],
		"isLastPage": true,
		"size": 3
	}`
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	reactions, err := client.ListCommitCommentReactions("MYPROJ", "my-svc", "abc123", 99)
	require.NoError(t, err)

	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/commits/abc123/comments/99/reactions", gotPath)
	require.Len(t, reactions, 2)

	thumbs := reactions[0]
	assert.Equal(t, "thumbs_up", thumbs.Emoji)
	require.Len(t, thumbs.Users, 2)
	assert.Equal(t, "alice", thumbs.Users[0].Slug)
	assert.Equal(t, "bob", thumbs.Users[1].Slug)

	heart := reactions[1]
	assert.Equal(t, "heart", heart.Emoji)
	require.Len(t, heart.Users, 1)
	assert.Equal(t, "carol", heart.Users[0].Slug)
}

func TestServerClient_ListCommitCommentReactions_EmptyReturnsNil(t *testing.T) {
	t.Parallel()
	const body = `{"values": [], "isLastPage": true, "size": 0}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	reactions, err := client.ListCommitCommentReactions("MYPROJ", "my-svc", "abc123", 99)
	require.NoError(t, err)
	assert.Empty(t, reactions)
}

func TestServerClient_ListCommitCommentReactions_NormalisesEmojiValues(t *testing.T) {
	t.Parallel()
	const body = `{
		"values": [
			{"emoticon": {"value": "thumbsup"},  "user": {"slug": "alice"}},
			{"emoticon": {"value": ":smile:"},   "user": {"slug": "bob"}},
			{"emoticon": {"value": "thumbs_up"}, "user": {"slug": "carol"}}
		],
		"isLastPage": true,
		"size": 3
	}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	reactions, err := client.ListCommitCommentReactions("MYPROJ", "my-svc", "abc123", 99)
	require.NoError(t, err)

	require.Len(t, reactions, 2, "thumbsup/thumbs_up group + smile(laugh)")
	assert.Equal(t, "thumbs_up", reactions[0].Emoji)
	assert.Len(t, reactions[0].Users, 2, "alice and carol both map to thumbs_up")
	assert.Equal(t, "laugh", reactions[1].Emoji)
}

// ── AddCommitCommentReaction ──────────────────────────────────────────────────

func TestServerClient_AddCommitCommentReaction_PostsCorrectPayload(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	err := client.AddCommitCommentReaction("MYPROJ", "my-svc", "abc123", 99, "thumbs_up")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/commits/abc123/comments/99/reactions", gotPath)
	emoticon, ok := gotBody["emoticon"].(map[string]any)
	require.True(t, ok, "expected emoticon object in body, got %#v", gotBody)
	assert.Equal(t, "thumbs_up", emoticon["value"])
}

// ── RemoveCommitCommentReaction ───────────────────────────────────────────────

func TestServerClient_RemoveCommitCommentReaction_DeletesCorrectURL(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	err := client.RemoveCommitCommentReaction("MYPROJ", "my-svc", "abc123", 99, "heart")
	require.NoError(t, err)

	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/commits/abc123/comments/99/reactions/heart", gotPath)
}

// ── CommitCommentReactor interface guard ──────────────────────────────────────

func TestServerClient_ImplementsCommitCommentReactor(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	reactor, err := backend.AsCommitCommentReactor(client, "git.example.com")
	require.NoError(t, err)
	require.NotNil(t, reactor)
}
