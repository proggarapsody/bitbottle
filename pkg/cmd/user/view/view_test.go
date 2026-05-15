package view_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/user/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "git.example.com:\n  oauth_token: tok\n"
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdView_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdView_RejectsArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"unexpected"})
	require.Error(t, cmd.Execute())
}

func TestView_PrintsSlugAndName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "jdoe", DisplayName: "Jane Doe"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "jdoe")
	assert.Contains(t, got, "Jane Doe")
}

func TestView_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "jdoe", DisplayName: "Jane Doe"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "jdoe")
	assert.Contains(t, got, "Jane Doe")
	assert.Contains(t, got, "slug")
	assert.Contains(t, got, "name")
}

func TestView_GetCurrentUserError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, errors.New("500 internal server error")
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestView_CloudBackend_Succeeds(t *testing.T) {
	// Ensure UserGetter is satisfied by the FakeClient when using a Cloud backend.
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice Cloud"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Cloud")
}
