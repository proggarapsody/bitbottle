package revoke_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo/revoke"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestRepoRevoke_User(t *testing.T) {
	t.Parallel()
	var gotProject, gotSlug string
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		RevokeRepoPermissionFn: func(_ context.Context, project, slug string, subject backend.PermissionSubject) error {
			gotProject, gotSlug = project, slug
			gotSubject = subject
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--user", "carol"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", gotProject)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "user", gotSubject.Kind)
	assert.Equal(t, "carol", gotSubject.Slug)
	assert.Contains(t, out.String(), "Revoked")
}

func TestRepoRevoke_Group(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		RevokeRepoPermissionFn: func(_ context.Context, _, _ string, subject backend.PermissionSubject) error {
			gotSubject = subject
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--group", "devs"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "group", gotSubject.Kind)
	assert.Equal(t, "devs", gotSubject.Name)
}

func TestRepoRevoke_RequiresExactlyOneSubjectFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-repo"}) // no --user or --group
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of --user or --group")
}

func TestRepoRevoke_InvalidArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "--user", "alice"}) // missing /REPO
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PROJECT/REPO")
}
