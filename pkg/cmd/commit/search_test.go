package commit_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitSearch_PrintsTable(t *testing.T) {
	t.Parallel()

	now := time.Now()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.Commit{
				{
					Hash:      "abc1234def567890",
					Message:   "feat: add search endpoint",
					Author:    backend.User{Slug: "alice"},
					Timestamp: now.Add(-1 * time.Hour),
				},
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc1234d") // 8-char truncation in non-TTY
	assert.Contains(t, got, "feat: add search endpoint")
	assert.Contains(t, got, "alice")
}

func TestCommitSearch_QueryFlag(t *testing.T) {
	t.Parallel()

	var gotOpts backend.CommitSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			gotOpts = opts
			return nil, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--query", "fix"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "fix", gotOpts.Query)
}

func TestCommitSearch_AuthorFlag(t *testing.T) {
	t.Parallel()

	var gotOpts backend.CommitSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			gotOpts = opts
			return nil, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--author", "alice"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "alice", gotOpts.Author)
}

func TestCommitSearch_SinceUntilFlags(t *testing.T) {
	t.Parallel()

	var gotOpts backend.CommitSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			gotOpts = opts
			return nil, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--since", "2026-01-01", "--until", "2026-06-01"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "2026-01-01", gotOpts.Since)
	assert.Equal(t, "2026-06-01", gotOpts.Until)
}

func TestCommitSearch_LimitFlag(t *testing.T) {
	t.Parallel()

	var gotOpts backend.CommitSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			gotOpts = opts
			return nil, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--limit", "5"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 5, gotOpts.Limit)
}

func TestCommitSearch_InvalidLimit(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--limit", "0"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--limit")
}

func TestCommitSearch_JSONOutput(t *testing.T) {
	t.Parallel()

	now := time.Now()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			return []backend.Commit{
				{
					Hash:      "abc1234def567890",
					Message:   "feat: add search endpoint",
					Author:    backend.User{Slug: "alice"},
					Timestamp: now.Add(-1 * time.Hour),
				},
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	var results []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &results))
	require.Len(t, results, 1)
	assert.Equal(t, "abc1234def567890", results[0]["hash"])
}
