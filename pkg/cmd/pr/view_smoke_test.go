package pr_test

// Bug: pr view prints p.Author.Slug for the "Author:" line. For Bitbucket
// Cloud, Author.Slug is the AccountID (a UUID like "abc123-uuid"), not a
// human-readable name. DisplayName should be preferred when available.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRView_Author_PrefersDisplayNameOverSlug(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{
				ID:    1,
				Title: "My PR",
				State: "OPEN",
				Author: backend.User{
					Slug:        "abc123-uuid", // Cloud AccountID
					DisplayName: "Alice Smith",
				},
				FromBranch: "feat/x",
				ToBranch:   "main",
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"1"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Alice Smith", "should show DisplayName, not UUID")
	assert.NotContains(t, got, "abc123-uuid", "should not show raw AccountID slug")
}

func TestPRView_Author_FallsBackToSlugWhenNoDisplayName(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{
				ID:    1,
				Title: "My PR",
				State: "OPEN",
				Author: backend.User{
					Slug:        "alice",
					DisplayName: "", // no display name
				},
				FromBranch: "feat/x",
				ToBranch:   "main",
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"1"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "alice", "should fall back to Slug when DisplayName is empty")
}
