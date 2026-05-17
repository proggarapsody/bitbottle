package pr_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// Each test captures the exact JSON field names emitted by the command so
// that any accidental rename of a Field.Name value causes a loud failure
// instead of silently breaking scripted consumers.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/pr/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			return []backend.PullRequest{
				{
					ID:         1,
					Title:      "Fix auth bug",
					State:      "OPEN",
					Draft:      false,
					FromBranch: "fix/auth",
					ToBranch:   "main",
					Author:     backend.User{Slug: "alice"},
					WebURL:     "https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/1",
				},
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/pr-list", out.String())
}

func TestPRView_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{
				ID:          42,
				Title:       "Add feature X",
				State:       "OPEN",
				Draft:       false,
				Description: "This PR adds feature X.",
				AutoMerge:   nil,
				FromBranch:  "feat/x",
				ToBranch:    "main",
				Author:      backend.User{Slug: "bob"},
				WebURL:      "https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/42",
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	format.RegisterOutputFlags(cmd)
	// pr view with --json uses prFieldsWithDescription (id passed as arg, repo resolved from remote)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/pr-view", out.String())
}
