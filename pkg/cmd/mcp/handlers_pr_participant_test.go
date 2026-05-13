package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListPRParticipants_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRParticipantsFn: func(ns, slug string, prID int) ([]backend.PRParticipant, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 42, prID)
			return []backend.PRParticipant{
				{
					User:     backend.User{Slug: "alice", DisplayName: "Alice Smith"},
					Role:     "AUTHOR",
					Approved: false,
					State:    "",
				},
				{
					User:     backend.User{Slug: "bob", DisplayName: "Bob Jones"},
					Role:     "REVIEWER",
					Approved: true,
					State:    "APPROVED",
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRParticipants(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "")
	assertJSONContains(t, result, "AUTHOR", "")
}

func TestListPRParticipants_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRParticipants(context.Background(), makeReq(map[string]any{
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestListPRParticipants_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRParticipants(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestListPRParticipants_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRParticipantsFn: func(ns, slug string, prID int) ([]backend.PRParticipant, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRParticipants(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
