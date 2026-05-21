package watcher_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/repo/watcher/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/watcher"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoWatcherList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			return []backend.User{
				{Slug: "alice", DisplayName: "Alice Smith"},
			}, nil
		},
	}

	// watcher list already calls format.RegisterOutputFlags in production — do NOT call it here.
	f, out, _ := newRepoFactory(t, fake)
	cmd := watcher.NewCmdWatcherList(f)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/repo-watcher-list", out.String())
}
