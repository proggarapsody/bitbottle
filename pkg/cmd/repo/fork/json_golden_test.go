package fork_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/repo/fork/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/fork"
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
	cmd := fork.NewCmdForkList(f)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/repo-fork-list", out.String())
}
