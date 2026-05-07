package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// noDefaultReviewersFake wraps backend.Client without satisfying
// DefaultReviewersResolver. Used to simulate a Bitbucket Cloud invocation.
type noDefaultReviewersFake struct {
	backend.Client
}

func TestPRCreate_AutoFetchesAndAppliesDefaultReviewers(t *testing.T) {
	t.Parallel()
	var captured backend.CreatePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		DefaultReviewersFn: func(ns, slug, fromBranch, toBranch string) ([]backend.User, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "feat/my-feature", fromBranch)
			assert.Equal(t, "main", toBranch)
			return []backend.User{{Slug: "alice"}, {Slug: "bob"}}, nil
		},
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1)), nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "T", "--base", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, []string{"alice", "bob"}, captured.Reviewers,
		"server backend with default reviewers must auto-apply them on create")
}

func TestPRCreate_ExplicitReviewersCombineWithDefaults(t *testing.T) {
	t.Parallel()
	var captured backend.CreatePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		DefaultReviewersFn: func(ns, slug, fromBranch, toBranch string) ([]backend.User, error) {
			return []backend.User{{Slug: "alice"}, {Slug: "bob"}}, nil
		},
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1)), nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{
		"--title", "T", "--base", "main",
		"--reviewer", "carol",
		"--reviewer", "alice", // also in defaults — must dedupe
	})
	require.NoError(t, cmd.Execute())
	// Explicit-first ordering: carol (explicit), alice (explicit, also in defaults), bob (defaults).
	assert.Equal(t, []string{"carol", "alice", "bob"}, captured.Reviewers)
}

func TestPRCreate_NoDefaultReviewersFlag_SkipsAutoFetch(t *testing.T) {
	t.Parallel()
	var captured backend.CreatePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		// DefaultReviewersFn intentionally unset — calling it would Fatalf.
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1)), nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{
		"--title", "T", "--base", "main",
		"--no-default-reviewers",
		"--reviewer", "carol",
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, []string{"carol"}, captured.Reviewers)
}

func TestPRCreate_CloudBackend_SkipsDefaultReviewers(t *testing.T) {
	t.Parallel()
	// Cloud backend = doesn't satisfy DefaultReviewersResolver. Auto-fetch
	// must silently skip; explicit --reviewer values still apply.
	var captured backend.CreatePRInput
	innerFake := &testhelpers.FakeClient{
		T: t,
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1)), nil
		},
	}
	f, _, _ := newPRFactory(t, noDefaultReviewersFake{Client: innerFake}, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "T", "--base", "main", "--reviewer", "alice"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, []string{"alice"}, captured.Reviewers)
}

func TestPRCreate_DefaultReviewersError_DegradesToWarning(t *testing.T) {
	t.Parallel()
	var captured backend.CreatePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		DefaultReviewersFn: func(ns, slug, fromBranch, toBranch string) ([]backend.User, error) {
			return nil, errors.New("403 forbidden")
		},
		CreatePRFn: func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1)), nil
		},
	}
	f, _, errOut := newPRFactory(t, fake, newPRCreateRunner())
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "T", "--base", "main", "--reviewer", "alice"})
	require.NoError(t, cmd.Execute(),
		"a default-reviewers fetch error must NOT fail the create flow")
	assert.Equal(t, []string{"alice"}, captured.Reviewers)
	assert.Contains(t, errOut.String(), "default reviewers")
	assert.Contains(t, errOut.String(), "warning")
}
