package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configcmd "github.com/proggarapsody/bitbottle/pkg/cmd/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func TestConfigSetGet_RoundTrips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: dir,
	})

	setCmd := configcmd.NewCmdConfig(f)
	setCmd.SetArgs([]string{"set", "editor", "nvim"})
	require.NoError(t, setCmd.Execute())

	// File written
	assert.FileExists(t, filepath.Join(dir, "config.yml"))

	getCmd := configcmd.NewCmdConfig(f)
	getCmd.SetArgs([]string{"get", "editor"})
	require.NoError(t, getCmd.Execute())
	assert.Equal(t, "nvim\n", out.String())
}

func TestConfigGet_UnknownKey_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: t.TempDir(),
	})
	cmd := configcmd.NewCmdConfig(f)
	cmd.SetArgs([]string{"get", "notakey"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "notakey")
}

func TestConfigSet_RejectsUnknownKey(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: t.TempDir(),
	})
	cmd := configcmd.NewCmdConfig(f)
	cmd.SetArgs([]string{"set", "secret_url", "x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown key")
}

func TestConfigList_PrintsAllSetValues(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: dir,
	})

	for _, args := range [][]string{
		{"set", "editor", "vim"},
		{"set", "pager", "less"},
	} {
		cmd := configcmd.NewCmdConfig(f)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
	}
	out.Reset()

	list := configcmd.NewCmdConfig(f)
	list.SetArgs([]string{"list"})
	require.NoError(t, list.Execute())
	got := out.String()
	assert.Contains(t, got, "editor=vim")
	assert.Contains(t, got, "pager=less")
}

func TestConfigSet_PerHostOverride(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: dir,
	})

	set := configcmd.NewCmdConfig(f)
	set.SetArgs([]string{"set", "git_protocol", "https", "--host", "bb.example.com"})
	require.NoError(t, set.Execute())

	get := configcmd.NewCmdConfig(f)
	get.SetArgs([]string{"get", "git_protocol", "--host", "bb.example.com"})
	require.NoError(t, get.Execute())
	assert.Equal(t, "https\n", out.String())
}
