package commit_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/commit/...

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitLog_JSONGolden(t *testing.T) {
	t.Parallel()

	// Use a fixed timestamp so the golden file is deterministic.
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitsFn: func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
			return []backend.Commit{
				{
					Hash:      "abc1234def567890abcd",
					Message:   "feat: add new feature",
					Author:    backend.User{Slug: "alice"},
					Timestamp: fixedTime,
				},
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitLog(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/commit-log", out.String())
}
