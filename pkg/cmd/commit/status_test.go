package commit_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitStatus_RendersTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitStatusesFn: func(ns, slug, hash string) ([]backend.CommitStatus, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "abc1234", hash)
			return []backend.CommitStatus{
				{Key: "build-1", State: "SUCCESSFUL", Name: "CI", Description: "all green", URL: "https://ci/1"},
				{Key: "lint-1", State: "FAILED", Name: "Lint", Description: "errors", URL: "https://ci/2"},
			}, nil
		},
	}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitStatus(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "build-1")
	assert.Contains(t, got, "SUCCESSFUL")
	assert.Contains(t, got, "lint-1")
	assert.Contains(t, got, "FAILED")
}

func TestCommitStatus_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitStatusesFn: func(ns, slug, hash string) ([]backend.CommitStatus, error) {
			return []backend.CommitStatus{
				{Key: "k1", State: "SUCCESSFUL", Name: "n", Description: "d", URL: "u"},
			}, nil
		},
	}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitStatus(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234", "--json", "key,state"})
	require.NoError(t, cmd.Execute())

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "k1", rows[0]["key"])
	assert.Equal(t, "SUCCESSFUL", rows[0]["state"])
}
