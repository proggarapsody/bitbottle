package repo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoTree_RequiresRefFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoTree(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoTree_RootListing_CallsBackendWithEmptyPath(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotRef, gotPath string
	fake := &testhelpers.FakeClient{
		T: t,
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			gotNS, gotSlug, gotRef, gotPath = ns, slug, ref, path
			return []backend.TreeEntry{
				{Path: "README.md", Type: "file", Size: 100},
				{Path: "cmd", Type: "dir"},
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoTree(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "--ref", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-svc", gotSlug)
	assert.Equal(t, "main", gotRef)
	assert.Equal(t, "", gotPath, "no PATH arg means root listing — empty string, not '/'")
	got := out.String()
	// Defaults to TSV (non-TTY iostream). Header is suppressed in non-TTY mode.
	assert.True(t, strings.Contains(got, "README.md") && strings.Contains(got, "cmd"),
		"output should list both entries; got: %q", got)
}

func TestRepoTree_NestedPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	fake := &testhelpers.FakeClient{
		T: t,
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			gotPath = path
			return nil, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoTree(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "cmd/foo", "--ref", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "cmd/foo", gotPath)
}

func TestRepoTree_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			return []backend.TreeEntry{
				{Path: "README.md", Type: "file", Size: 100},
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoTree(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "--ref", "main", "--json", "path,type,size"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"path":"README.md"`)
	assert.Contains(t, got, `"type":"file"`)
	assert.Contains(t, got, `"size":100`)
}

func TestRepoTree_JQFilter(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListTreeFn: func(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
			return []backend.TreeEntry{
				{Path: "README.md", Type: "file"},
				{Path: "cmd", Type: "dir"},
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoTree(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "--ref", "main", "--json", "path,type", "--jq", ".[].path"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "README.md")
	assert.Contains(t, got, "cmd")
}
