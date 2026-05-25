package put_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/file/put"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const repoConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

func newPutFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	return f, out, errOut
}

func TestRepoFilePut_RequiresBranchFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--message", "Update", "--content", "x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch")
}

func TestRepoFilePut_RequiresMessageFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--content", "x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message")
}

func TestRepoFilePut_RequiresContentOrContentFile(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "content")
}

func TestRepoFilePut_ContentAndContentFileMutuallyExclusive(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPutFactory(t, fake)
	dir := t.TempDir()
	cf := filepath.Join(dir, "f.txt")
	require.NoError(t, os.WriteFile(cf, []byte("x"), 0o644))
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update", "--content", "x", "--content-file", cf})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestRepoFilePut_CallsBackendWithParsedArgs(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotPath string
	var gotIn backend.PutFileInput
	fake := &testhelpers.FakeClient{
		T: t,
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			gotNS, gotSlug, gotPath = ns, slug, path
			gotIn = in
			return nil
		},
	}
	f, out, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update README", "--content", "# Hello"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-svc", gotSlug)
	assert.Equal(t, "README.md", gotPath)
	assert.Equal(t, "main", gotIn.Branch)
	assert.Equal(t, "Update README", gotIn.Message)
	assert.Equal(t, "# Hello", gotIn.Content)
	assert.Contains(t, out.String(), "README.md")
}

func TestRepoFilePut_ContentFileFlag(t *testing.T) {
	t.Parallel()
	var gotContent string
	fake := &testhelpers.FakeClient{
		T: t,
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			gotContent = in.Content
			return nil
		},
	}
	dir := t.TempDir()
	cf := filepath.Join(dir, "hello.txt")
	require.NoError(t, os.WriteFile(cf, []byte("file content here"), 0o644))

	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"hello.txt", "MYPROJ/my-svc", "--branch", "main", "--message", "Add hello", "--content-file", cf})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "file content here", gotContent)
}

func TestRepoFilePut_SourceCommitFlag(t *testing.T) {
	t.Parallel()
	var gotSourceCommit string
	fake := &testhelpers.FakeClient{
		T: t,
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			gotSourceCommit = in.SourceCommit
			return nil
		},
	}
	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update", "--content", "x", "--source-commit", "deadbeef"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "deadbeef", gotSourceCommit)
}

func TestRepoFilePut_BackendErrorPropagates(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("conflict")
	fake := &testhelpers.FakeClient{
		T: t,
		PutFileFn: func(ns, slug, path string, in backend.PutFileInput) error {
			return wantErr
		},
	}
	f, _, _ := newPutFactory(t, fake)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update", "--content", "x"})
	err := cmd.Execute()
	require.ErrorIs(t, err, wantErr)
}

func TestRepoFilePut_UnsupportedOnBackend(t *testing.T) {
	t.Parallel()
	// Plain Client without SourceWriter — wraps FakeClient
	type noWriteFake struct{ backend.Client }
	wrapped := noWriteFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := newPutFactory(t, wrapped)
	cmd := put.NewCmdFilePut(f)
	cmd.SetArgs([]string{"README.md", "MYPROJ/my-svc", "--branch", "main", "--message", "Update", "--content", "x"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
