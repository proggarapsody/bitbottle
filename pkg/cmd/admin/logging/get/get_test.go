package get_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/logging/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestLoggingGet_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetLoggingConfigFn: func() (backend.LoggingConfig, error) {
			return backend.LoggingConfig{Level: "INFO", Async: false}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Level: INFO")
	assert.Contains(t, out.String(), "Async: false")
}

func TestLoggingGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetLoggingConfigFn: func() (backend.LoggingConfig, error) {
			return backend.LoggingConfig{Level: "DEBUG", Async: true}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"level":"DEBUG"`)
	assert.Contains(t, out.String(), `"async":true`)
}

func TestLoggingGet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
