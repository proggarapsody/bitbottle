package defaultreviewer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/project/defaultreviewer"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := defaultreviewer.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdList_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := defaultreviewer.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsReviewers(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListProjectDefaultReviewersFn: func(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "PROJ", projectKey)
			assert.Equal(t, 50, limit)
			return []backend.ProjectDefaultReviewer{
				{AccountID: "abc123", DisplayName: "Alice", Nickname: "alice"},
				{AccountID: "def456", DisplayName: "Bob", Nickname: "bob"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := defaultreviewer.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc123")
	assert.Contains(t, got, "Alice")
	assert.Contains(t, got, "def456")
	assert.Contains(t, got, "Bob")
}

func TestList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListProjectDefaultReviewersFn: func(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error) {
			return []backend.ProjectDefaultReviewer{
				{AccountID: "abc123", DisplayName: "Alice", Nickname: "alice"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := defaultreviewer.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc123")
	assert.Contains(t, got, "Alice")
}

// noDefaultReviewerFake wraps backend.Client without satisfying WorkspaceProjectDefaultReviewerClient.
type noDefaultReviewerFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noDefaultReviewerFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := defaultreviewer.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
