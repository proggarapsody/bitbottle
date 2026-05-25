package get_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/mail/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestMailGet_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetMailServerConfigFn: func() (backend.MailServerConfig, error) {
			return backend.MailServerConfig{
				Hostname:      "smtp.example.com",
				Port:          25,
				Protocol:      "smtp",
				UseStartTLS:   false,
				Username:      "mailer",
				SenderAddress: "no-reply@example.com",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdMailGet(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "smtp.example.com")
	assert.Contains(t, out.String(), "no-reply@example.com")
}

func TestMailGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetMailServerConfigFn: func() (backend.MailServerConfig, error) {
			return backend.MailServerConfig{
				Hostname: "smtp.example.com",
				Port:     465,
				Protocol: "smtps",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdMailGet(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"Hostname":"smtp.example.com"`)
	assert.Contains(t, out.String(), `"Port":465`)
	// Password must never appear in JSON output (json:"-" tag)
	assert.NotContains(t, out.String(), "Password")
}

func TestMailGet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := get.NewCmdMailGet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
