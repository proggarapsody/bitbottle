package prsettings_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	prsettings "github.com/proggarapsody/bitbottle/pkg/cmd/repo/pr-settings"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const testConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

func newFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: testConfig})
	factorytest.UseBackend(f, fake)
	return f, out, errOut
}

func TestRepoPRSettingsGet_PrintsTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			return backend.RepoPRSettings{
				RequiredApprovers:        2,
				RequiredAllApprovers:     true,
				RequiredAllTasksComplete: false,
				RequiredSuccessfulBuilds: 1,
				MergeStrategy:            "no-ff",
				AllowedStrategies:        []string{"no-ff", "squash"},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"get", "MYPROJ/my-repo"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "no-ff")
}

func TestRepoPRSettingsGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{
				RequiredApprovers: 3,
				MergeStrategy:     "squash",
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"get", "MYPROJ/my-repo", "--json"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, `"requiredApprovers"`)
	assert.Contains(t, output, "3")
}

func TestRepoPRSettingsGet_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{}, errors.New("500 internal")
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"get", "MYPROJ/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
