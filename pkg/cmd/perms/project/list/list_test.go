package list_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func newFakePermsClient(fn func(ctx context.Context, project string) ([]backend.PermissionGrant, error)) *testhelpers.FakeClient {
	return &testhelpers.FakeClient{
		ListProjectPermissionsFn: fn,
	}
}

func TestProjectList_PrintsRows(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, project string) ([]backend.PermissionGrant, error) {
			assert.Equal(t, "MYPROJ", project)
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "alice", DisplayName: "Alice A"}, Permission: "PROJECT_ADMIN"},
				{Subject: backend.PermissionSubject{Kind: "group", Name: "devs"}, Permission: "PROJECT_WRITE"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "PROJECT_ADMIN")
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "devs")
}

func TestProjectList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, project string) ([]backend.PermissionGrant, error) {
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "bob"}, Permission: "PROJECT_READ"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"permission":"PROJECT_READ"`)
	assert.Contains(t, got, `"name":"bob"`)
}

func TestProjectList_CloudBackend_ReturnsUnsupported(t *testing.T) {
	t.Parallel()
	const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: cloud\n"

	type noPermsClient struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noPermsClient{Client: &testhelpers.FakeClient{}})

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Server / Data Center only")
}
