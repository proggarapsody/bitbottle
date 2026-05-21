package stop_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/stop"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdStop_TTY_NoConfirmNeeded(t *testing.T) {
	t.Parallel()

	var gotOpts *stop.Options
	runF := func(opts *stop.Options) error {
		gotOpts = opts
		return nil
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	// Simulate TTY stdout so --confirm is not required in runF path.
	f.IOStreams.IsStdoutTTY = func() bool { return true }

	cmd := stop.NewCmdStop(f, runF)
	cmd.SetArgs([]string{"abc-uuid-123"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotOpts)
	assert.Equal(t, "abc-uuid-123", gotOpts.Args[0])
	assert.False(t, gotOpts.Confirm)
}

func TestNewCmdStop_NonTTY_WithoutConfirm_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "bitbucket.org:\n  oauth_token: tok\n",
	})
	// Non-TTY (factorytest default: IsStdoutTTY returns false).
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}

	cmd := stop.NewCmdStop(f, nil)
	cmd.SetArgs([]string{"abc-uuid-123", "myws/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestNewCmdStop_NonTTY_WithConfirm_Succeeds(t *testing.T) {
	t.Parallel()

	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		StopPipelineFn: func(ws, slug, uuid string) error {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc-uuid-123", uuid)
			called = true
			return nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "bitbucket.org:\n  oauth_token: tok\n",
	})
	factorytest.UseBackend(f, fake)
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}

	cmd := stop.NewCmdStop(f, nil)
	cmd.SetArgs([]string{"abc-uuid-123", "myws/my-repo", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.True(t, strings.Contains(out.String(), "abc-uuid-123"))
}
