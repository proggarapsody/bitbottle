package issue_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/issue/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/issue"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestIssueList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListIssuesFn: func(ns, slug, state string, limit int) ([]backend.Issue, error) {
			return []backend.Issue{
				{
					ID:       7,
					Title:    "Example bug",
					State:    "open",
					Kind:     "bug",
					Priority: "major",
					Reporter: backend.User{Slug: "alice"},
					WebURL:   "https://bitbucket.org/acme/repo/issues/7",
				},
			}, nil
		},
	}

	f, out, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme/repo", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/issue-list", out.String())
}

func TestIssueView_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetIssueFn: func(ns, slug string, id int) (backend.Issue, error) {
			return backend.Issue{
				ID:       42,
				Title:    "Sample issue",
				State:    "open",
				Kind:     "bug",
				Priority: "minor",
				Reporter: backend.User{Slug: "bob"},
				Content:  "This is the issue body.",
				WebURL:   "https://bitbucket.org/acme/repo/issues/42",
			}, nil
		},
	}

	f, out, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueView(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme/repo", "42", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/issue-view", out.String())
}
