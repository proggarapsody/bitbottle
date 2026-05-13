package set_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/logging/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestLoggingSet_Level_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.LoggingConfigInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error {
			gotIn = in
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--level", "WARN"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "WARN", gotIn.Level)
	assert.False(t, gotIn.Persistent)
	assert.Contains(t, out.String(), "runtime-only")
}

func TestLoggingSet_Persistent_PrintsNote(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T:                  t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error { return nil },
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--level", "ERROR", "--persistent"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "persistent")
}

func TestLoggingSet_Async_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.LoggingConfigInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--async"})
	require.NoError(t, cmd.Execute())
	assert.True(t, gotIn.Async)
}

func TestLoggingSet_NoFlags_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of --level or --async")
}

func TestLoggingSet_InvalidLevel_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--level", "warn"}) // lowercase — invalid
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "case-sensitive")
}

func TestLoggingSet_PermissionError_PrintsHint(t *testing.T) {
	t.Parallel()
	permErr := &backend.DomainError{Kind: backend.ErrPermission, Message: "forbidden"}
	fake := &testhelpers.FakeClient{
		T:                  t,
		SetLoggingConfigFn: func(in backend.LoggingConfigInput) error { return permErr },
	}
	f, _, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--level", "DEBUG"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "SYS_ADMIN")
}

func TestLoggingSet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--level", "INFO"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
