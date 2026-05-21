package edit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/edit"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const repoConfig = "bb.example.com:\n  oauth_token: tok\n"

func TestNewCmdEdit_NoFlagsErrors(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})
	cmd := edit.NewCmdEdit(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestNewCmdEdit_MutuallyExclusiveIssues(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})
	cmd := edit.NewCmdEdit(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--enable-issues", "--disable-issues"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestNewCmdEdit_MutuallyExclusiveWiki(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{})
	cmd := edit.NewCmdEdit(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--enable-wiki", "--disable-wiki"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestNewCmdEdit_DescriptionCallsBackend(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotIn backend.EditRepoInput
	fake := &testhelpers.FakeClient{
		EditRepoFn: func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
			gotNS, gotSlug, gotIn = ns, slug, in
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	cmd := edit.NewCmdEdit(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--description", "new desc"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	require.NotNil(t, gotIn.Description)
	assert.Equal(t, "new desc", *gotIn.Description)
	assert.Contains(t, out.String(), "MYPROJ/my-service")
	assert.Contains(t, out.String(), "updated")
}

func TestNewCmdEdit_EnableIssuesSetsHasIssuesTrue(t *testing.T) {
	t.Parallel()
	var gotIn backend.EditRepoInput
	fake := &testhelpers.FakeClient{
		EditRepoFn: func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	cmd := edit.NewCmdEdit(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--enable-issues"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotIn.HasIssues)
	assert.True(t, *gotIn.HasIssues)
}

func TestNewCmdEdit_DisableWikiSetsHasWikiFalse(t *testing.T) {
	t.Parallel()
	var gotIn backend.EditRepoInput
	fake := &testhelpers.FakeClient{
		EditRepoFn: func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	cmd := edit.NewCmdEdit(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--disable-wiki"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotIn.HasWiki)
	assert.False(t, *gotIn.HasWiki)
}
