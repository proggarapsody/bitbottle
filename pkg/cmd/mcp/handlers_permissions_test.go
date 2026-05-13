package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func fakePermsHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	return newHandlersWithFake(t, singleHostConfig, fake)
}

// ── listProjectPermissions ────────────────────────────────────────────────────

func TestMCP_ListProjectPermissions_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListProjectPermissionsFn: func(_ context.Context, project string) ([]backend.PermissionGrant, error) {
			assert.Equal(t, "MYPROJ", project)
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "alice"}, Permission: "PROJECT_ADMIN"},
			}, nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.listProjectPermissions(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "PROJECT_ADMIN")
}

func TestMCP_ListProjectPermissions_MissingProject(t *testing.T) {
	t.Parallel()
	h := fakePermsHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.listProjectPermissions(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

// ── grantProjectPermission ────────────────────────────────────────────────────

func TestMCP_GrantProjectPermission_OK(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	var gotPerm string
	fake := &testhelpers.FakeClient{T: t,
		GrantProjectPermissionFn: func(_ context.Context, project string, subject backend.PermissionSubject, perm string) error {
			assert.Equal(t, "MYPROJ", project)
			gotSubject = subject
			gotPerm = perm
			return nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.grantProjectPermission(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "permission": "PROJECT_WRITE", "user": "alice",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "granted", "")
	assert.Equal(t, "user", gotSubject.Kind)
	assert.Equal(t, "alice", gotSubject.Slug)
	assert.Equal(t, "PROJECT_WRITE", gotPerm)
}

func TestMCP_GrantProjectPermission_BothSubjectFlags(t *testing.T) {
	t.Parallel()
	h := fakePermsHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.grantProjectPermission(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "permission": "PROJECT_READ", "user": "u", "group": "g",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "exactly one")
}

// ── revokeProjectPermission ───────────────────────────────────────────────────

func TestMCP_RevokeProjectPermission_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		RevokeProjectPermissionFn: func(_ context.Context, project string, subject backend.PermissionSubject) error {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "group", subject.Kind)
			assert.Equal(t, "devs", subject.Name)
			called = true
			return nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.revokeProjectPermission(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "group": "devs",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "revoked", "")
	assert.True(t, called)
}

// ── listRepoPermissions ───────────────────────────────────────────────────────

func TestMCP_ListRepoPermissions_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListRepoPermissionsFn: func(_ context.Context, project, slug string) ([]backend.PermissionGrant, error) {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "my-repo", slug)
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "carol"}, Permission: "REPO_ADMIN"},
			}, nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.listRepoPermissions(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "carol", "REPO_ADMIN")
}

// ── grantRepoPermission ───────────────────────────────────────────────────────

func TestMCP_GrantRepoPermission_OK(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{T: t,
		GrantRepoPermissionFn: func(_ context.Context, project, slug string, subject backend.PermissionSubject, perm string) error {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "REPO_WRITE", perm)
			gotSubject = subject
			return nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.grantRepoPermission(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "permission": "REPO_WRITE", "group": "qa",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "granted", "")
	assert.Equal(t, "group", gotSubject.Kind)
	assert.Equal(t, "qa", gotSubject.Name)
}

// ── revokeRepoPermission ──────────────────────────────────────────────────────

func TestMCP_RevokeRepoPermission_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		RevokeRepoPermissionFn: func(_ context.Context, project, slug string, subject backend.PermissionSubject) error {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "user", subject.Kind)
			assert.Equal(t, "bob", subject.Slug)
			called = true
			return nil
		},
	}
	h := fakePermsHandlers(t, fake)
	result, err := h.revokeRepoPermission(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "user": "bob",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "revoked", "")
	assert.True(t, called)
}

// ── noPermsClient returns unsupported ────────────────────────────────────────

// noPermsClientWrapper wraps backend.Client without satisfying PermissionsClient,
// simulating a Cloud backend invocation.
type noPermsClientWrapper struct{ backend.Client }

func TestMCP_ListProjectPermissions_Unsupported(t *testing.T) {
	t.Parallel()
	// Build a handlers instance whose backend does NOT implement PermissionsClient.
	// noPermsClientWrapper embeds backend.Client (interface), which excludes the
	// PermissionsClient methods — so AsPermissionsClient will return ErrUnsupportedOnHost.
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	wrapper := noPermsClientWrapper{Client: &testhelpers.FakeClient{T: t}}
	factorytest.UseBackend(f, wrapper)
	h := newHandlers(f)

	result, err := h.listProjectPermissions(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ",
	}))
	require.NoError(t, err) // handler returns nil error; wraps in error result
	require.NotNil(t, result)
	assert.True(t, result.IsError, "expected error result for unsupported backend")
	// The error envelope should contain the host.unsupported code and
	// a mention of "permissions" (the feature name).
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, text.Text, "host.unsupported")
	assert.Contains(t, text.Text, "permissions")
}
