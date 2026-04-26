package tag_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestTagDelete_ConfirmationPrompt_Abort(t *testing.T) {
	t.Parallel()

	var deleted bool
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteTagFn: func(ns, slug, name string) error {
			deleted = true
			return nil
		},
	}

	out := &strings.Builder{}
	ios := &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader("n\n")),
		Out:         out,
		ErrOut:      io.Discard,
		IsStdoutTTY: func() bool { return true },
		IsStderrTTY: func() bool { return false },
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
		IOStreams:        ios,
	})
	cmd := tag.NewCmdTagDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0"})
	require.NoError(t, cmd.Execute())

	assert.False(t, deleted, "expected delete not to be called after abort")
	assert.Contains(t, out.String(), "Aborted")
}

func TestTagDelete_ConfirmationPrompt_Confirm(t *testing.T) {
	t.Parallel()

	var deletedName string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteTagFn: func(ns, slug, name string) error {
			deletedName = name
			return nil
		},
	}

	ios := &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader("y\n")),
		Out:         io.Discard,
		ErrOut:      io.Discard,
		IsStdoutTTY: func() bool { return true },
		IsStderrTTY: func() bool { return false },
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
		IOStreams:        ios,
	})
	cmd := tag.NewCmdTagDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "v1.0.0", deletedName)
}

func TestTagDelete_SkipsPromptWithConfirmFlag(t *testing.T) {
	t.Parallel()

	var deletedName string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteTagFn: func(ns, slug, name string) error {
			deletedName = name
			return nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "v1.0.0", deletedName)
}

func TestTagDelete_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		DeleteTagFn: func(ns, slug, name string) error {
			return errors.New("tag not found")
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag not found")
}
