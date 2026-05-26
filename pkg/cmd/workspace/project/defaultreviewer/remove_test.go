package defaultreviewer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/project/defaultreviewer"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRemove_RemovesReviewer(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotAccountID string
	fake := &testhelpers.FakeClient{
		T: t,
		RemoveProjectDefaultReviewerFn: func(workspace, projectKey, accountID string) error {
			gotWS, gotKey, gotAccountID = workspace, projectKey, accountID
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := defaultreviewer.NewCmdRemove(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "abc123", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "abc123", gotAccountID)
	assert.Contains(t, out.String(), "Removed")
	assert.Contains(t, out.String(), "abc123")
}

func TestRemove_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := defaultreviewer.NewCmdRemove(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "abc123"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestRemove_RequiresUserFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := defaultreviewer.NewCmdRemove(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRemove_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noDefaultReviewerFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := defaultreviewer.NewCmdRemove(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "abc123", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
