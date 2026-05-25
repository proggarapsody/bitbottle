package rerun_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/rerun"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRerun_ParsesArgs(t *testing.T) {
	t.Parallel()

	var gotOpts *rerun.Options
	runF := func(opts *rerun.Options) error {
		gotOpts = opts
		return nil
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := rerun.NewCmdRerun(f, runF)
	cmd.SetArgs([]string{"abc-uuid-123", "myws/my-repo"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotOpts)
	assert.Equal(t, "abc-uuid-123", gotOpts.Args[0])
	assert.Equal(t, "myws/my-repo", gotOpts.Args[1])
}

func TestNewCmdRerun_Success_TTY(t *testing.T) {
	t.Parallel()

	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		RerunPipelineFn: func(ns, slug, srcUUID string) (backend.Pipeline, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc-uuid-123", srcUUID)
			called = true
			return backend.Pipeline{
				UUID:        "newuuid",
				BuildNumber: 77,
				WebURL:      "https://bitbucket.org/myws/my-repo/pipelines/results/77",
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "bitbucket.org:\n  oauth_token: tok\n",
	})
	factorytest.UseBackend(f, fake)
	f.IOStreams.IsStdoutTTY = func() bool { return true }
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}

	cmd := rerun.NewCmdRerun(f, nil)
	cmd.SetArgs([]string{"abc-uuid-123", "myws/my-repo"})
	require.NoError(t, cmd.Execute())

	assert.True(t, called)
	assert.True(t, strings.Contains(out.String(), "#77"))
}

func TestNewCmdRerun_Success_NonTTY(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		RerunPipelineFn: func(ns, slug, srcUUID string) (backend.Pipeline, error) {
			return backend.Pipeline{
				UUID:        "newuuid",
				BuildNumber: 88,
				WebURL:      "https://bitbucket.org/myws/my-repo/pipelines/results/88",
			}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "bitbucket.org:\n  oauth_token: tok\n",
	})
	factorytest.UseBackend(f, fake)
	// Default: non-TTY
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}

	cmd := rerun.NewCmdRerun(f, nil)
	cmd.SetArgs([]string{"abc-uuid-123", "myws/my-repo"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "88")
	assert.Contains(t, output, "pipelines/results/88")
}

func TestNewCmdRerun_ClientError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		RerunPipelineFn: func(ns, slug, srcUUID string) (backend.Pipeline, error) {
			return backend.Pipeline{}, errors.New("pipeline not found")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: "bitbucket.org:\n  oauth_token: tok\n",
	})
	factorytest.UseBackend(f, fake)
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "my-repo"}, nil
	}

	cmd := rerun.NewCmdRerun(f, nil)
	cmd.SetArgs([]string{"nonexistent-uuid", "myws/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline not found")
}
