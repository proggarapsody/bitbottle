package use_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/profile/use"
)

func prepareStore(t *testing.T) *profiles.Store {
	t.Helper()
	store := profiles.New(t.TempDir())
	require.NoError(t, store.Load())
	store.Set("work", profiles.Profile{
		Hostname:      "git.work.com",
		Token:         "work-token",
		User:          "alice",
		AuthUser:      "alice@work.com",
		SkipTLSVerify: true,
		BackendType:   "server",
		GitProtocol:   "https",
	})
	return store
}

func TestUseRun_Happy(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := prepareStore(t)
	factorytest.UseProfiles(f, store)

	cmd := use.NewCmdUse(f, nil)
	cmd.SetArgs([]string{"work"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Switched to profile work (git.work.com)")

	// Verify that the host config was written.
	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("git.work.com")
	require.True(t, ok)
	assert.Equal(t, "work-token", hc.OAuthToken)
	assert.Equal(t, "alice", hc.User)
	assert.Equal(t, "alice@work.com", hc.AuthUser)
	assert.True(t, hc.SkipTLSVerify)
	assert.Equal(t, "server", hc.BackendType)
	assert.Equal(t, "https", hc.GitProtocol)
}

func TestUseRun_ProfileNotFound(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := use.NewCmdUse(f, nil)
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `profile "nonexistent" not found`)
}

func TestUseRun_AppliesMinimalProfile(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	require.NoError(t, store.Load())
	store.Set("minimal", profiles.Profile{
		Hostname: "bitbucket.org",
		Token:    "bb-tok",
	})
	factorytest.UseProfiles(f, store)

	// Pre-configure the host with an existing user to prove merge semantics.
	cfg, err := f.Config()
	require.NoError(t, err)
	cfg.Set("bitbucket.org", config.HostConfig{User: "preserved-user"})
	require.NoError(t, cfg.Save())

	cmd := use.NewCmdUse(f, nil)
	cmd.SetArgs([]string{"minimal"})
	require.NoError(t, cmd.Execute())

	cfg, err = f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bitbucket.org")
	require.True(t, ok)
	// Profile token is applied.
	assert.Equal(t, "bb-tok", hc.OAuthToken)
	// Pre-existing user is preserved (merge, not replace).
	assert.Equal(t, "preserved-user", hc.User)
}
