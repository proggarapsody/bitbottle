package list_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestRepoList_PrintsRows(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListRepoPermissionsFn: func(_ context.Context, project, slug string) ([]backend.PermissionGrant, error) {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "my-repo", slug)
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "carol"}, Permission: "REPO_ADMIN"},
				{Subject: backend.PermissionSubject{Kind: "group", Name: "qa team"}, Permission: "REPO_READ"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "REPO_ADMIN")
	assert.Contains(t, got, "carol")
	assert.Contains(t, got, "qa team")
}

func TestRepoList_InvalidArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ"}) // missing /REPO
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PROJECT/REPO")
}

func TestRepoList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		ListRepoPermissionsFn: func(_ context.Context, _, _ string) ([]backend.PermissionGrant, error) {
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "bob"}, Permission: "REPO_WRITE"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"permission":"REPO_WRITE"`)
	assert.Contains(t, got, `"name":"bob"`)
}
