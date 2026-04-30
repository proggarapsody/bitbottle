package commit_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const commitConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestCommitLog_PrintsTable(t *testing.T) {
	t.Parallel()

	now := time.Now()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			return []backend.Commit{
				{
					Hash:      "abc1234def567890",
					Message:   "feat: add new feature",
					Author:    backend.User{Slug: "alice"},
					Timestamp: now.Add(-1 * time.Hour),
					WebURL:    "https://example.com/commits/abc1234def567890",
				},
				{
					Hash:      "deadbeef01234567",
					Message:   "fix: resolve crash on startup",
					Author:    backend.User{Slug: "bob"},
					Timestamp: now.Add(-2 * time.Hour),
					WebURL:    "https://example.com/commits/deadbeef01234567",
				},
				{
					Hash:      "cafebabe12345678",
					Message:   "chore: update dependencies",
					Author:    backend.User{DisplayName: "Carol Smith"},
					Timestamp: now.Add(-3 * time.Hour),
					WebURL:    "https://example.com/commits/cafebabe12345678",
				},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitLog(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	// Hashes should be truncated to 7 chars in pipe mode (non-TTY)
	assert.Contains(t, got, "abc1234")
	assert.Contains(t, got, "deadbee")
	assert.Contains(t, got, "cafebab")
	assert.Contains(t, got, "feat: add new feature")
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "bob")
	// Carol has no Slug, should fall back to DisplayName
	assert.Contains(t, got, "Carol Smith")
}

func TestCommitLog_BranchFlag(t *testing.T) {
	t.Parallel()

	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			gotBranch = branch
			return nil, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitLog(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "feat/x"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "feat/x", gotBranch)
}

func TestCommitLog_DefaultBranch(t *testing.T) {
	t.Parallel()

	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{Name: "master", IsDefault: false},
				{Name: "dev", IsDefault: true},
			}, nil
		},
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			gotBranch = branch
			return nil, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitLog(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "dev", gotBranch)
}

func TestCommitLog_DefaultBranch_FallsBackToMain(t *testing.T) {
	t.Parallel()

	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "feature-x", IsDefault: false}}, nil
		},
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			gotBranch = branch
			return nil, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitLog(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "main", gotBranch)
}

func TestCommitLog_JSONOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			return []backend.Commit{
				{
					Hash:      "abc1234def567890",
					Message:   "feat: add new feature",
					Author:    backend.User{Slug: "alice"},
					Timestamp: time.Now().Add(-1 * time.Hour),
					WebURL:    "https://example.com/commits/abc1234def567890",
				},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitLog(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main", "--json", "hash,message"})
	require.NoError(t, cmd.Execute())

	var results []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &results))
	require.Len(t, results, 1)

	// JSON output should contain the full hash, not a truncated version.
	assert.Equal(t, "abc1234def567890", results[0]["hash"])
	assert.Equal(t, "feat: add new feature", results[0]["message"])
}
