package revoke_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/revoke"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestProjectRevoke_User(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		RevokeProjectPermissionFn: func(_ context.Context, project string, subject backend.PermissionSubject) error {
			gotSubject = subject
			assert.Equal(t, "MYPROJ", project)
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "--user", "alice"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "user", gotSubject.Kind)
	assert.Equal(t, "alice", gotSubject.Slug)
	assert.Contains(t, out.String(), "Revoked")
}

func TestProjectRevoke_Group(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		RevokeProjectPermissionFn: func(_ context.Context, _ string, subject backend.PermissionSubject) error {
			gotSubject = subject
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "--group", "devs"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "group", gotSubject.Kind)
	assert.Equal(t, "devs", gotSubject.Name)
}

func TestProjectRevoke_RequiresExactlyOneSubjectFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})

	cmd := revoke.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"MYPROJ"}) // no --user or --group
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of --user or --group")
}
