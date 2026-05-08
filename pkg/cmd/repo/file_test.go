package repo_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoFileGet_RequiresRefFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFileGet(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "main.go"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ref")
}

func TestRepoFileGet_CallsBackendWithParsedArgs(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotRef, gotPath string
	fake := &testhelpers.FakeClient{
		T: t,
		GetFileContentFn: func(ns, slug, ref, path string) ([]byte, error) {
			gotNS, gotSlug, gotRef, gotPath = ns, slug, ref, path
			return []byte("package main\n"), nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFileGet(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "cmd/main.go", "--ref", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-svc", gotSlug)
	assert.Equal(t, "main", gotRef)
	assert.Equal(t, "cmd/main.go", gotPath)
	assert.Equal(t, "package main\n", out.String())
}

func TestRepoFileGet_WritesToOutFile(t *testing.T) {
	t.Parallel()
	binary := []byte{0xff, 0xd8, 0xff, 0xe0}
	fake := &testhelpers.FakeClient{
		T: t,
		GetFileContentFn: func(ns, slug, ref, path string) ([]byte, error) {
			return binary, nil
		},
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "logo.jpg")
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFileGet(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "logo.jpg", "--ref", "main", "--out", out})
	require.NoError(t, cmd.Execute())
	got, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, binary, got, "binary content must round-trip via --out")
}

func TestRepoFileGet_BackendErrorPropagates(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("boom")
	fake := &testhelpers.FakeClient{
		T: t,
		GetFileContentFn: func(ns, slug, ref, path string) ([]byte, error) {
			return nil, wantErr
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFileGet(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "main.go", "--ref", "main"})
	err := cmd.Execute()
	require.ErrorIs(t, err, wantErr)
}

func TestRepoFileGet_RequiresPath(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoFileGet(f)
	cmd.SetArgs([]string{"MYPROJ/my-svc", "", "--ref", "main"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoFile_HasGetSubcommand(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newRepoFactory(t, fake)
	parent := repo.NewCmdRepoFile(f)
	var have []string
	for _, c := range parent.Commands() {
		have = append(have, c.Name())
	}
	assert.Contains(t, have, "get")
}

// Compile-time guard: catch backend.SourceReader signature drift at the
// command-level binding (not just the adapter level).
var _ = func() backend.SourceReader { return (*testhelpers.FakeClient)(nil) }
