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

// ── list_groups ───────────────────────────────────────────────────────────────

func TestListGroups_Success(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			gotFilter = filter
			return []backend.Group{{Name: "developers"}, {Name: "admins"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listGroups(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "", gotFilter)
	assertJSONContains(t, result, "developers", "")
}

func TestListGroups_WithFilter(t *testing.T) {
	t.Parallel()
	var gotFilter string
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			gotFilter = filter
			return []backend.Group{{Name: "devs"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listGroups(context.Background(), makeReq(map[string]any{"filter": "dev"}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "dev", gotFilter)
}

func TestListGroups_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupsFn: func(filter string, limit int) ([]backend.Group, error) {
			return nil, errors.New("403 forbidden")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listGroups(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "403")
}

// ── create_group ──────────────────────────────────────────────────────────────

func TestCreateGroup_Success(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		CreateGroupFn: func(name string) (backend.Group, error) {
			gotName = name
			return backend.Group{Name: name}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createGroup(context.Background(), makeReq(map[string]any{"name": "newgroup"}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "newgroup", gotName)
	assertJSONContains(t, result, "newgroup", "")
}

func TestCreateGroup_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createGroup(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

// ── delete_group ──────────────────────────────────────────────────────────────

func TestDeleteGroup_Success(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteGroupFn: func(name string) error {
			gotName = name
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteGroup(context.Background(), makeReq(map[string]any{"name": "oldgroup"}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "oldgroup", gotName)
}

func TestDeleteGroup_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteGroup(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

// ── list_group_members ────────────────────────────────────────────────────────

func TestListGroupMembers_Success(t *testing.T) {
	t.Parallel()
	var gotGroup string
	fake := &testhelpers.FakeClient{
		T: t,
		ListGroupMembersFn: func(groupName string, limit int) ([]backend.GroupMember, error) {
			gotGroup = groupName
			return []backend.GroupMember{
				{Name: "alice", DisplayName: "Alice", EmailAddress: "alice@example.com"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listGroupMembers(context.Background(), makeReq(map[string]any{"group": "developers"}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "developers", gotGroup)
	assertJSONContains(t, result, "alice", "")
}

func TestListGroupMembers_MissingGroup(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listGroupMembers(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "group")
}

// ── add_group_member ──────────────────────────────────────────────────────────

func TestAddGroupMember_Success(t *testing.T) {
	t.Parallel()
	var gotGroup, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		AddGroupMemberFn: func(groupName, user string) error {
			gotGroup = groupName
			gotUser = user
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.addGroupMember(context.Background(), makeReq(map[string]any{
		"group": "developers",
		"user":  "alice",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "developers", gotGroup)
	assert.Equal(t, "alice", gotUser)
}

func TestAddGroupMember_MissingGroup(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addGroupMember(context.Background(), makeReq(map[string]any{"user": "alice"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "group")
}

func TestAddGroupMember_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addGroupMember(context.Background(), makeReq(map[string]any{"group": "developers"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}

// ── remove_group_member ───────────────────────────────────────────────────────

func TestRemoveGroupMember_Success(t *testing.T) {
	t.Parallel()
	var gotGroup, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		RemoveGroupMemberFn: func(groupName, user string) error {
			gotGroup = groupName
			gotUser = user
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.removeGroupMember(context.Background(), makeReq(map[string]any{
		"group": "developers",
		"user":  "alice",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "developers", gotGroup)
	assert.Equal(t, "alice", gotUser)
}

func TestRemoveGroupMember_MissingGroup(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeGroupMember(context.Background(), makeReq(map[string]any{"user": "alice"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "group")
}

func TestRemoveGroupMember_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeGroupMember(context.Background(), makeReq(map[string]any{"group": "developers"}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}
