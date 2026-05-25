package clear_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/clear"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBannerClear_WithConfirm_Succeeds(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		ClearBannerFn: func() error {
			called = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := clear.NewCmdBannerClear(f, nil)
	cmd.SetArgs([]string{"--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "cleared")
}

func TestBannerClear_NoConfirm_NoTTY_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := clear.NewCmdBannerClear(f, nil)
	// no --confirm flag; factory default is non-TTY
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required in non-interactive mode")
}

func TestBannerClear_TTY_UserConfirmsY(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		ClearBannerFn: func() error {
			called = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	f.IOStreams.IsStdoutTTY = func() bool { return true }
	f.IOStreams.In = io.NopCloser(strings.NewReader("y\n"))
	cmd := clear.NewCmdBannerClear(f, nil)
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "cleared")
}

func TestBannerClear_TTY_UserDeclines(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	f.IOStreams.IsStdoutTTY = func() bool { return true }
	f.IOStreams.In = io.NopCloser(strings.NewReader("n\n"))
	cmd := clear.NewCmdBannerClear(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Aborted")
}

func TestBannerClear_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := clear.NewCmdBannerClear(f, nil)
	cmd.SetArgs([]string{"--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
