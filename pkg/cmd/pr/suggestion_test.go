package pr_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	pr "github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const suggestionConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

func newSuggestionRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "ssh://git@bb.example.com:7999/myproj/my-service.git\n",
	})
}

// serverFakeClientSuggestion wraps FakeClient and also satisfies SuggestionApplier,
// simulating a Server/DC backend where suggestions are supported.
type serverFakeClientSuggestion struct {
	*testhelpers.FakeClient
	ApplySuggestionFn      func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error)
	GetSuggestionPreviewFn func(ns, slug string, prID, commentID int) (string, error)
}

func (s *serverFakeClientSuggestion) ApplySuggestion(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
	if s.ApplySuggestionFn != nil {
		return s.ApplySuggestionFn(ns, slug, prID, commentID, suggestionID)
	}
	s.T.Fatalf("unexpected call to serverFakeClientSuggestion.ApplySuggestion; set ApplySuggestionFn")
	return backend.SuggestionApplyResult{}, nil
}

func (s *serverFakeClientSuggestion) GetSuggestionPreview(ns, slug string, prID, commentID int) (string, error) {
	if s.GetSuggestionPreviewFn != nil {
		return s.GetSuggestionPreviewFn(ns, slug, prID, commentID)
	}
	s.T.Fatalf("unexpected call to serverFakeClientSuggestion.GetSuggestionPreview; set GetSuggestionPreviewFn")
	return "", nil
}

func newSuggestionFactory(t *testing.T, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: suggestionConfig})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}

// ── apply tests ───────────────────────────────────────────────────────────────

func TestPRSuggestionApply_OK(t *testing.T) {
	t.Parallel()
	var gotPRID, gotCommentID, gotSuggestionID int
	fake := &serverFakeClientSuggestion{
		FakeClient: &testhelpers.FakeClient{T: t},
		ApplySuggestionFn: func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
			gotPRID = prID
			gotCommentID = commentID
			gotSuggestionID = suggestionID
			return backend.SuggestionApplyResult{CommitHash: "deadbeef", CommitMessage: "Suggestion applied"}, nil
		},
	}
	f, out, _ := newSuggestionFactory(t, fake, newSuggestionRunner())
	cmd := pr.NewCmdSuggestion(f)
	cmd.SetArgs([]string{"apply", "42", "7", "1"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, 1, gotSuggestionID)
	assert.Contains(t, out.String(), "deadbeef")
}

func TestPRSuggestionApply_Preview(t *testing.T) {
	t.Parallel()
	fake := &serverFakeClientSuggestion{
		FakeClient: &testhelpers.FakeClient{T: t},
		GetSuggestionPreviewFn: func(ns, slug string, prID, commentID int) (string, error) {
			return "suggestion preview text", nil
		},
	}
	f, out, _ := newSuggestionFactory(t, fake, newSuggestionRunner())
	cmd := pr.NewCmdSuggestion(f)
	cmd.SetArgs([]string{"apply", "42", "7", "1", "--preview"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "suggestion preview text")
}

func TestPRSuggestionApply_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// FakeClient does NOT implement SuggestionApplier → Cloud-like backend.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newSuggestionFactory(t, fake, newSuggestionRunner())
	cmd := pr.NewCmdSuggestion(f)
	cmd.SetArgs([]string{"apply", "42", "7", "1"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, backend.ErrUnsupportedOnHost))
}

func TestPRSuggestionApply_MissingArgs(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newSuggestionFactory(t, fake, newSuggestionRunner())
	cmd := pr.NewCmdSuggestion(f)
	cmd.SetArgs([]string{"apply", "42", "7"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRSuggestionApply_BackendError(t *testing.T) {
	t.Parallel()
	fake := &serverFakeClientSuggestion{
		FakeClient: &testhelpers.FakeClient{T: t},
		ApplySuggestionFn: func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
			return backend.SuggestionApplyResult{}, errors.New("apply failed")
		},
	}
	f, _, _ := newSuggestionFactory(t, fake, newSuggestionRunner())
	cmd := pr.NewCmdSuggestion(f)
	cmd.SetArgs([]string{"apply", "42", "7", "1"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "apply failed")
}
