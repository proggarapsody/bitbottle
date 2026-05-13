package grant_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project/grant"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestProjectGrant_HappyPath_User(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	var gotPerm string
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, _ string) ([]backend.PermissionGrant, error) {
			return nil, nil // no existing grants
		},
		GrantProjectPermissionFn: func(_ context.Context, project string, subject backend.PermissionSubject, perm string) error {
			gotSubject = subject
			gotPerm = perm
			assert.Equal(t, "MYPROJ", project)
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "PROJECT_WRITE", "--user", "alice"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "user", gotSubject.Kind)
	assert.Equal(t, "alice", gotSubject.Slug)
	assert.Equal(t, "PROJECT_WRITE", gotPerm)
	assert.Contains(t, out.String(), "Granted")
}

func TestProjectGrant_HappyPath_Group(t *testing.T) {
	t.Parallel()
	var gotSubject backend.PermissionSubject
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, _ string) ([]backend.PermissionGrant, error) {
			return nil, nil
		},
		GrantProjectPermissionFn: func(_ context.Context, _ string, subject backend.PermissionSubject, _ string) error {
			gotSubject = subject
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "PROJECT_READ", "--group", "qa team"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "group", gotSubject.Kind)
	assert.Equal(t, "qa team", gotSubject.Name)
}

func TestProjectGrant_RequiresExactlyOneSubjectFlag(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"MYPROJ", "PROJECT_READ"},                                   // neither
		{"MYPROJ", "PROJECT_READ", "--user", "a", "--group", "g"},   // both
	}
	for _, args := range cases {
		f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
		factorytest.UseBackend(f, &testhelpers.FakeClient{})
		cmd := grant.NewCmdGrant(f, nil)
		cmd.SetArgs(args)
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one of --user or --group")
	}
}

func TestProjectGrant_DowngradeWarning_Confirmed(t *testing.T) {
	t.Parallel()
	var grantCalled bool
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, _ string) ([]backend.PermissionGrant, error) {
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "alice"}, Permission: "PROJECT_ADMIN"},
			}, nil
		},
		GrantProjectPermissionFn: func(_ context.Context, _ string, _ backend.PermissionSubject, _ string) error {
			grantCalled = true
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)
	// Inject "y\n" as stdin so the confirmation is accepted.
	f.IOStreams.In = io.NopCloser(strings.NewReader("y\n"))
	f.IOStreams.IsStdoutTTY = func() bool { return true }

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "PROJECT_READ", "--user", "alice"})
	require.NoError(t, cmd.Execute())

	assert.True(t, grantCalled, "grant must be called when downgrade confirmed")
	assert.Contains(t, out.String(), "downgrade")
}

func TestProjectGrant_DowngradeWarning_Rejected(t *testing.T) {
	t.Parallel()
	var grantCalled bool
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, _ string) ([]backend.PermissionGrant, error) {
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "alice"}, Permission: "PROJECT_ADMIN"},
			}, nil
		},
		GrantProjectPermissionFn: func(_ context.Context, _ string, _ backend.PermissionSubject, _ string) error {
			grantCalled = true
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)
	f.IOStreams.In = io.NopCloser(strings.NewReader("n\n"))
	f.IOStreams.IsStdoutTTY = func() bool { return true }

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "PROJECT_READ", "--user", "alice"})
	require.NoError(t, cmd.Execute())

	assert.False(t, grantCalled, "grant must NOT be called when downgrade rejected")
	assert.Contains(t, out.String(), "Aborted")
}

func TestProjectGrant_Force_SkipsDowngradePrompt(t *testing.T) {
	t.Parallel()
	var grantCalled bool
	fake := &testhelpers.FakeClient{
		ListProjectPermissionsFn: func(_ context.Context, _ string) ([]backend.PermissionGrant, error) {
			return []backend.PermissionGrant{
				{Subject: backend.PermissionSubject{Kind: "user", Slug: "alice"}, Permission: "PROJECT_ADMIN"},
			}, nil
		},
		GrantProjectPermissionFn: func(_ context.Context, _ string, _ backend.PermissionSubject, _ string) error {
			grantCalled = true
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig, BackendType: "server"})
	factorytest.UseBackend(f, fake)

	cmd := grant.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"MYPROJ", "PROJECT_READ", "--user", "alice", "--force"})
	require.NoError(t, cmd.Execute())

	assert.True(t, grantCalled)
}
