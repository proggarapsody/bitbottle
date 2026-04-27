package commit_test

// Bug: commit view always prints a "Web:" line even when c.WebURL is empty.
// Server commits never have a WebURL populated by the API, so the output
// incorrectly shows "Web:     " with a blank value.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitView_NoWebURL_DoesNotPrintWebLine(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCommitFn: func(ns, slug, hash string) (backend.Commit, error) {
			return backend.Commit{
				Hash:      "abc1234",
				Message:   "feat: something",
				Author:    backend.User{Slug: "alice"},
				Timestamp: time.Now(),
				WebURL:    "", // no URL — typical for Server commits
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234"})
	require.NoError(t, cmd.Execute())

	assert.NotContains(t, out.String(), "Web:", "should not print Web: line when URL is empty")
}

func TestCommitView_WithWebURL_PrintsWebLine(t *testing.T) {
	t.Parallel()

	url := "https://bb.example.com/projects/PROJ/repos/svc/commits/abc1234"
	fake := &testhelpers.FakeClient{
		T: t,
		GetCommitFn: func(ns, slug, hash string) (backend.Commit, error) {
			return backend.Commit{
				Hash:      "abc1234",
				Message:   "feat: something",
				Author:    backend.User{Slug: "alice"},
				Timestamp: time.Now(),
				WebURL:    url,
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Web:", "should print Web: line when URL is set")
	assert.Contains(t, out.String(), url)
}
