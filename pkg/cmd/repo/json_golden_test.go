package repo_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/repo/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoForkList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoForksFn: func(ns, slug string, limit int) ([]backend.Repository, error) {
			return []backend.Repository{
				{
					Namespace: "teamA",
					Slug:      "my-service-fork",
					Name:      "my-service-fork",
				},
			}, nil
		},
	}

	// fork-list already calls format.RegisterOutputFlags in production — do NOT call it here.
	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoForkList(f)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/repo-fork-list", out.String())
}

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
	cmd := repo.NewCmdRepoWatcherList(f)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/repo-watcher-list", out.String())
}
