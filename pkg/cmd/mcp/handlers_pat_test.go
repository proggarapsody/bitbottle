package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// singleHostConfigWithUser is a hosts.yml config that includes user slug,
// required for PAT operations (which need the userSlug from config).
const singleHostConfigWithUser = "git.example.com:\n  user: alice\n  oauth_token: tok\n"

func newPATHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfigWithUser})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

// ── list_pats ─────────────────────────────────────────────────────────────────

func TestMCP_ListPATs(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			assert.Equal(t, "alice", userSlug)
			return []backend.PAT{
				{ID: "1", Name: "CI Token", Permissions: []string{"REPO_READ"}, CreatedDate: time.Now()},
				{ID: "2", Name: "Dev Token", Permissions: []string{"REPO_WRITE"}, CreatedDate: time.Now()},
			}, nil
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.listPATs(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "CI Token", "")
	assertJSONContains(t, result, "Dev Token", "")
}

func TestMCP_ListPATs_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Code:    backend.CodeHostUnsupported,
				Message: "personal access token management is not supported on bitbucket.org (Bitbucket Server / Data Center only)",
			}
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.listPATs(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestMCP_ListPATs_DefaultLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	_, err := h.listPATs(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 50, gotLimit)
}

// ── create_pat ────────────────────────────────────────────────────────────────

func TestMCP_CreatePAT(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreatePATInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			assert.Equal(t, "alice", userSlug)
			gotInput = in
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "42", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-supersecret",
			}, nil
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{
		"hostname": "git.example.com",
		"name":     "CI Token",
		"scopes":   "repo:read,repo:write",
	})
	result, err := h.createPAT(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "BBDC-supersecret", "")
	assertJSONContains(t, result, "42", "")
	assert.Equal(t, "CI Token", gotInput.Name)
	assert.Contains(t, gotInput.Permissions, "REPO_READ")
	assert.Contains(t, gotInput.Permissions, "REPO_WRITE")
	assert.Nil(t, gotInput.ExpiryDays)
}

func TestMCP_CreatePAT_WithExpiry(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreatePATInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			gotInput = in
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "7", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-exp",
			}, nil
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{
		"hostname":   "git.example.com",
		"name":       "Expiring",
		"scopes":     "pr:read",
		"expires_in": float64(30),
	})
	result, err := h.createPAT(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	require.NotNil(t, gotInput.ExpiryDays)
	assert.Equal(t, 30, *gotInput.ExpiryDays)
}

func TestMCP_CreatePAT_MissingName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com", "scopes": "repo:read"})
	result, err := h.createPAT(context.Background(), req)
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestMCP_CreatePAT_InvalidScope(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{
		"hostname": "git.example.com",
		"name":     "T",
		"scopes":   "bad:scope",
	})
	result, err := h.createPAT(context.Background(), req)
	require.NoError(t, err)
	assertErrorResult(t, result, "unknown scope")
}

// ── revoke_pat ────────────────────────────────────────────────────────────────

func TestMCP_RevokePAT(t *testing.T) {
	t.Parallel()
	var gotUserSlug, gotTokenID string
	fake := &testhelpers.FakeClient{
		T: t,
		RevokePATFn: func(userSlug, tokenID string) error {
			gotUserSlug = userSlug
			gotTokenID = tokenID
			return nil
		},
	}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com", "token_id": "42"})
	result, err := h.revokePAT(context.Background(), req)
	require.NoError(t, err)
	assertJSONContains(t, result, "revoked", "42")
	assert.Equal(t, "alice", gotUserSlug)
	assert.Equal(t, "42", gotTokenID)
}

func TestMCP_RevokePAT_MissingTokenID(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	h := newPATHandlers(t, fake)
	req := makeReq(map[string]any{"hostname": "git.example.com"})
	result, err := h.revokePAT(context.Background(), req)
	require.NoError(t, err)
	assertErrorResult(t, result, "token_id")
}
