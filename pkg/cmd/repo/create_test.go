package repo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoCreate_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoCreate(f)
	assert.NotNil(t, cmd.Flag("project"))
	assert.NotNil(t, cmd.Flag("description"))
	assert.NotNil(t, cmd.Flag("private"))
}

func TestRepoCreate_CallsAPIAndPrintsSummary(t *testing.T) {
	t.Parallel()

	var capturedNS string
	var capturedIn backend.CreateRepoInput

	fake := &testhelpers.FakeClient{
		T: t,
		CreateRepoFn: func(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
			capturedNS = ns
			capturedIn = in
			return testhelpers.BackendRepoFactory(
				testhelpers.BackendRepoWithSlug("new-repo"),
				testhelpers.BackendRepoWithWebURL("https://bb.example.com/projects/MYPROJ/repos/new-repo/browse"),
			), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoCreate(f)
	cmd.SetArgs([]string{"new-repo", "--project", "MYPROJ"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", capturedNS)
	assert.Equal(t, "new-repo", capturedIn.Name)
	assert.False(t, capturedIn.Public, "default --private=true means Public should be false")
	assert.Contains(t, out.String(), "new-repo")
}

func TestRepoCreate_MissingProject_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := newRepoFactory(t, nil)
	cmd := repo.NewCmdRepoCreate(f)
	cmd.SetArgs([]string{"new-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
}

func TestRepoCreate_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("409 conflict")
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRepoFn: func(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
			return backend.Repository{}, apiErr
		},
	}

	f, _, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoCreate(f)
	cmd.SetArgs([]string{"new-repo", "--project", "MYPROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409 conflict")
}

func TestNewCmdRepoCreate_HasJSONAndJQFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoCreate(f)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestRepoCreate_JSON_EmitsObject(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		CreateRepoFn: func(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("new-repo")), nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoCreate(f)
	cmd.SetArgs([]string{"new-repo", "--project", "MYPROJ", "--json", "slug,namespace"})
	require.NoError(t, cmd.Execute())

	got := strings.TrimSpace(out.String())
	assert.True(t, strings.HasPrefix(got, "{"), "expected JSON object, got: %s", got)
	assert.Contains(t, got, `"slug":"new-repo"`)
}
