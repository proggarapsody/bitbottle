package pr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// newPRCreateRunner returns a FakeRunner seeded with the two git calls that
// pr create needs: remote get-url origin (resolveRepoRef) and
// rev-parse --abbrev-ref HEAD (current branch detection).
func newPRCreateRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/my-service.git\n"},
		testhelpers.RunResponse{Stdout: "feat/my-feature\n"},
	)
}

func TestNewCmdPRCreate_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRCreate(f)
	assert.NotNil(t, cmd.Flag("title"))
	assert.NotNil(t, cmd.Flag("body"))
	assert.NotNil(t, cmd.Flag("base"))
	assert.NotNil(t, cmd.Flag("draft"))
}

func TestPRCreate_WithFlags_CallsAPIAndPrints(t *testing.T) {
	t.Parallel()

	var capturedNS, capturedSlug string
	var capturedIn backend.CreatePRInput

	fake := &testhelpers.FakeClient{
		T: t,
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			capturedNS = ns
			capturedSlug = slug
			capturedIn = in
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithTitle("My PR"),
				testhelpers.BackendPRWithWebURL("https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/42"),
			), nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "My PR", "--base", "main"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", capturedNS)
	assert.Equal(t, "my-service", capturedSlug)
	assert.Equal(t, "My PR", capturedIn.Title)
	assert.Equal(t, "main", capturedIn.ToBranch)
	assert.Contains(t, out.String(), "42")
}

func TestPRCreate_MissingTitle_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := newPRFactory(t, nil, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--base", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestPRCreate_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("409 conflict")
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			return backend.PullRequest{}, apiErr
		},
	}

	f, _, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "My PR", "--base", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409 conflict")
}

func TestNewCmdPRCreate_HasJSONAndJQFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRCreate(f)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestPRCreate_JSON_EmitsObject(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(99),
				testhelpers.BackendPRWithTitle("New PR"),
			), nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "New PR", "--base", "main", "--json", "id,title"})
	require.NoError(t, cmd.Execute())

	got := strings.TrimSpace(out.String())
	assert.True(t, strings.HasPrefix(got, "{"), "expected JSON object, got: %s", got)
	assert.Contains(t, got, `"id":99`)
	assert.Contains(t, got, `"title":"New PR"`)
}
