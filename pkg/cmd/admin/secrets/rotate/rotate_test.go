package rotate_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/secrets/rotate"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "git.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

func TestRotate_WithConfirm_Succeeds(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		RotateSecretsFn: func() error {
			called = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := rotate.NewCmdRotate(f, nil)
	cmd.SetArgs([]string{"--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "Secrets rotated")
}

func TestRotate_NoConfirm_NoTTY_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := rotate.NewCmdRotate(f, nil)
	// no --confirm flag; factory default is non-TTY
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required in non-interactive mode")
}

func TestRotate_PermissionError_PrintsHint(t *testing.T) {
	t.Parallel()
	permErr := &backend.DomainError{Kind: backend.ErrPermission, Message: "forbidden"}
	fake := &testhelpers.FakeClient{
		T: t,
		RotateSecretsFn: func() error {
			return permErr
		},
	}
	f, _, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := rotate.NewCmdRotate(f, nil)
	cmd.SetArgs([]string{"--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "SYS_ADMIN")
}

func TestRotate_UnsupportedOnCloud_ReturnsError(t *testing.T) {
	t.Parallel()
	// FakeClient does NOT implement AdminClient — simulate Cloud backend
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := rotate.NewCmdRotate(f, nil)
	cmd.SetArgs([]string{"--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}

func TestRotate_TTY_UserConfirmsY(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		RotateSecretsFn: func() error {
			called = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	// Simulate TTY
	f.IOStreams.IsStdoutTTY = func() bool { return true }
	f.IOStreams.In = io.NopCloser(strings.NewReader("y\n"))
	cmd := rotate.NewCmdRotate(f, nil)
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "Secrets rotated")
}

func TestRotate_TTY_UserDeclines(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	f.IOStreams.IsStdoutTTY = func() bool { return true }
	f.IOStreams.In = io.NopCloser(strings.NewReader("n\n"))
	cmd := rotate.NewCmdRotate(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Aborted")
}
