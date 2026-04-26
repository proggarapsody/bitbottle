package commit_test

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

func TestCommitView_PrintsDetail(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		GetCommitFn: func(ns, slug, hash string) (backend.Commit, error) {
			return backend.Commit{
				Hash:      "abc1234def567890abcdef1234567890abcdef12",
				Message:   "feat: implement commit view command",
				Author:    backend.User{Slug: "alice"},
				Timestamp: ts,
				WebURL:    "https://bitbucket.org/myworkspace/my-service/commits/abc1234def567890",
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
	})
	cmd := commit.NewCmdCommitView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234def567890abcdef1234567890abcdef12"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc1234def567890abcdef1234567890abcdef12")
	assert.Contains(t, got, "feat: implement commit view command")
	assert.Contains(t, got, "alice")
}

func TestCommitView_WebFlag(t *testing.T) {
	t.Parallel()

	webURL := "https://bitbucket.org/myworkspace/my-service/commits/abc1234def567890"
	fake := &testhelpers.FakeClient{
		T: t,
		GetCommitFn: func(ns, slug, hash string) (backend.Commit, error) {
			return backend.Commit{
				Hash:      "abc1234def567890",
				Message:   "feat: something",
				Author:    backend.User{Slug: "alice"},
				Timestamp: time.Now(),
				WebURL:    webURL,
			}, nil
		},
	}

	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   commitConfig,
		BackendOverride: fake,
		Browser:         browser,
	})
	cmd := commit.NewCmdCommitView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc1234def567890", "--web"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Equal(t, webURL, browser.URLs[0])
}
