package commit_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitFiles_RendersTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitFilesFn: func(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "abc1234", hash)
			return []backend.DiffStatEntry{
				{Path: "foo.go", Status: "modified", Additions: 5, Deletions: 2},
				{Path: "bar.go", Status: "added", Additions: 10, Deletions: 0},
				{Path: "baz.go", Status: "deleted", Additions: 0, Deletions: 3},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitFiles(f)
	cmd.SetArgs([]string{"abc1234", "myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "foo.go")
	assert.Contains(t, got, "modified")
	assert.Contains(t, got, "bar.go")
	assert.Contains(t, got, "added")
	assert.Contains(t, got, "baz.go")
	assert.Contains(t, got, "deleted")
}

func TestCommitFiles_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitFilesFn: func(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
			return []backend.DiffStatEntry{
				{Path: "foo.go", Status: "modified", Additions: 5, Deletions: 2},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitFiles(f)
	cmd.SetArgs([]string{"abc1234", "myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "foo.go", rows[0]["path"])
	assert.Equal(t, "modified", rows[0]["status"])
	assert.EqualValues(t, 5, rows[0]["additions"])
	assert.EqualValues(t, 2, rows[0]["deletions"])
}

func TestCommitFiles_HashOnly_UsesBaseRepo(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitFilesFn: func(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "deadbeef", hash)
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.DiffStatEntry{}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitFiles(f)
	cmd.SetArgs([]string{"deadbeef"})
	require.NoError(t, cmd.Execute())
}
