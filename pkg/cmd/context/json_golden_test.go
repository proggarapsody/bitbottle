package context_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/context/...

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	contextcmd "github.com/proggarapsody/bitbottle/pkg/cmd/context"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestContext_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(_, _ string, _ int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "main", IsDefault: true}}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice Smith"}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "feat/x"},
		testhelpers.RunResponse{Err: errors.New("no upstream")},
	)
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/context", out.String())
}
