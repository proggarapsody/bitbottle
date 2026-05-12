package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRMerge_HasAutoFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := newPRFactory(t, &testhelpers.FakeClient{T: t}, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	assert.NotNil(t, cmd.Flag("auto"))
	assert.NotNil(t, cmd.Flag("auto-off"))
	assert.NotNil(t, cmd.Flag("rebase"))
}

func TestPRMerge_AutoFlag_EnablesAutoMerge(t *testing.T) {
	t.Parallel()

	var gotNS, gotSlug, gotStrategy string
	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		EnableAutoMergeFn: func(ns, slug string, id int, strategy string) error {
			gotNS, gotSlug, gotID, gotStrategy = ns, slug, id, strategy
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--auto"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, 42, gotID)
	assert.Equal(t, "merge", gotStrategy)
	assert.Contains(t, out.String(), "auto-merge")
	assert.Contains(t, out.String(), "#42")
}

func TestPRMerge_AutoSquash_SendsSquashStrategy(t *testing.T) {
	t.Parallel()

	var gotStrategy string
	fake := &testhelpers.FakeClient{
		T: t,
		EnableAutoMergeFn: func(ns, slug string, id int, strategy string) error {
			gotStrategy = strategy
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--auto", "--squash"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "squash", gotStrategy)
}

func TestPRMerge_AutoRebase_SendsRebaseStrategy(t *testing.T) {
	t.Parallel()

	var gotStrategy string
	fake := &testhelpers.FakeClient{
		T: t,
		EnableAutoMergeFn: func(ns, slug string, id int, strategy string) error {
			gotStrategy = strategy
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--auto", "--rebase"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "rebase", gotStrategy)
}

func TestPRMerge_AutoOff_DisablesAutoMerge(t *testing.T) {
	t.Parallel()

	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		DisableAutoMergeFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--auto-off"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 42, gotID)
	assert.Contains(t, out.String(), "Cancelled auto-merge")
}

func TestPRMerge_AutoAndAutoOff_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := newPRFactory(t, &testhelpers.FakeClient{T: t}, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--auto", "--auto-off"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--auto")
	assert.Contains(t, err.Error(), "--auto-off")
}

// TestPRMerge_AutoMergeQueued_NonTTY_DoesNotPrompt tests that when the PR has
// auto-merge enabled and we're not on a TTY, the command proceeds without
// prompting (non-interactive path: auto-merge is silently cancelled and
// the immediate merge runs).
func TestPRMerge_AutoMergeQueued_NonTTY_DoesNotPrompt(t *testing.T) {
	t.Parallel()

	mergeCalled := false
	disableCalled := false
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			p := testhelpers.BackendPRFactory()
			p.AutoMerge = &backend.AutoMergeState{Enabled: true, Strategy: "merge"}
			return p, nil
		},
		DisableAutoMergeFn: func(ns, slug string, id int) error {
			disableCalled = true
			return nil
		},
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			mergeCalled = true
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithState("MERGED")), nil
		},
	}
	// Default test factory uses a non-TTY IOStreams.
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	assert.True(t, disableCalled, "auto-merge must be cancelled on non-TTY")
	assert.True(t, mergeCalled, "merge must proceed after cancel")
}
