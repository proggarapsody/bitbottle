package grant_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/grant"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestRepoGrant_HappyPath_User(t *testing.T) {
	t.Parallel()
	var gotProject, gotSlug string
	var gotSubject backend.PermissionSubject
	var gotPerm string
	fake := &testhelpers.FakeClient{
		ListRepoPermissionsFn: func(_ context.Context, _, _ string) ([]backend.PermissionGrant, error) {
			return nil, nil
		},
		GrantRepoPermissionFn: func(_ context.Context, project, slug string, subject backend.PermissionSubject, perm string) error {
			gotProject, gotSlug = project, slug
			gotSubject = subject
			gotPerm = perm
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "REPO_WRITE", "--user", "carol"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", gotProject)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "user", gotSubject.Kind)
	assert.Equal(t, "carol", gotSubject.Slug)
	assert.Equal(t, "REPO_WRITE", gotPerm)
	assert.Contains(t, out.String(), "Granted")
}

func TestRepoGrant_HappyPath_Group(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		ListRepoPermissionsFn: func(_ context.Context, _, _ string) ([]backend.PermissionGrant, error) {
			return nil, nil
		},
		GrantRepoPermissionFn: func(_ context.Context, _, _ string, subject backend.PermissionSubject, _ string) error {
			gotSubject = subject
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "REPO_READ", "--group", "qa team"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "group", gotSubject.Kind)
	assert.Equal(t, "qa team", gotSubject.Name)
}

func TestRepoGrant_RequiresExactlyOneSubjectFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "REPO_READ"}) // no --user or --group
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of --user or --group")
}

func TestRepoGrant_InvalidRepoArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "REPO_READ", "--user", "alice"}) // missing /REPO
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PROJECT/REPO")
}
