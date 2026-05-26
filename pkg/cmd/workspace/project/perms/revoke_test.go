package perms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/project/perms"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRevoke_RevokesUserPermission(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotSlug string
	var gotIsGroup bool
	fake := &testhelpers.FakeClient{
		T: t,
		RevokeWorkspaceProjectPermFn: func(workspace, projectKey, subjectSlug string, isGroup bool) error {
			gotWS, gotKey, gotSlug, gotIsGroup = workspace, projectKey, subjectSlug, isGroup
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "alice", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "alice", gotSlug)
	assert.False(t, gotIsGroup)
	assert.Contains(t, out.String(), "Revoked")
}

func TestRevoke_RevokesGroupPermission(t *testing.T) {
	t.Parallel()
	var gotSlug string
	var gotIsGroup bool
	fake := &testhelpers.FakeClient{
		T: t,
		RevokeWorkspaceProjectPermFn: func(workspace, projectKey, subjectSlug string, isGroup bool) error {
			gotSlug, gotIsGroup = subjectSlug, isGroup
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--group", "devs", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "devs", gotSlug)
	assert.True(t, gotIsGroup)
}

func TestRevoke_RequiresConfirmWhenNonTTY(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	// factorytest defaults to non-TTY
	cmd := perms.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "alice"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestRevoke_RequiresUserOrGroup(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := perms.NewCmdRevoke(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--user or --group")
}
