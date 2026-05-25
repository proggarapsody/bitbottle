package set_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/mail/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestMailSet_CallsBackendWithArgs(t *testing.T) {
	t.Parallel()
	var gotCfg backend.MailServerConfig
	fake := &testhelpers.FakeClient{
		T: t,
		SetMailServerConfigFn: func(in backend.MailServerConfig) error {
			gotCfg = in
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdMailSet(f, nil)
	cmd.SetArgs([]string{
		"--mail-hostname", "smtp.example.com",
		"--port", "465",
		"--protocol", "smtps",
		"--sender", "bot@example.com",
		"--username", "mailer",
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "smtp.example.com", gotCfg.Hostname)
	assert.Equal(t, 465, gotCfg.Port)
	assert.Equal(t, "smtps", gotCfg.Protocol)
	assert.Equal(t, "bot@example.com", gotCfg.SenderAddress)
	assert.Equal(t, "mailer", gotCfg.Username)
	assert.Contains(t, out.String(), "updated")
}

func TestMailSet_RequiresMailHostname(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdMailSet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mail-hostname")
}

func TestMailSet_PasswordWarning(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetMailServerConfigFn: func(in backend.MailServerConfig) error {
			return nil
		},
	}
	f, _, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdMailSet(f, nil)
	cmd.SetArgs([]string{
		"--mail-hostname", "smtp.example.com",
		"--password", "s3cr3t",
	})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, errOut.String(), "warning")
	assert.Contains(t, errOut.String(), "process list")
}

func TestMailSet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := set.NewCmdMailSet(f, nil)
	cmd.SetArgs([]string{"--mail-hostname", "smtp.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
