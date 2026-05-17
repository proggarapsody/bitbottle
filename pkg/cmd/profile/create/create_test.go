package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/profile/create"
)

func TestCreateRun_Happy(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work", "--hostname", "git.work.com", "--token", "tok123"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Profile work created (host: git.work.com)")

	p, ok := store.Get("work")
	require.True(t, ok)
	assert.Equal(t, "git.work.com", p.Hostname)
	assert.Equal(t, "tok123", p.Token)
}

func TestCreateRun_AllFlags(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"work",
		"--hostname", "git.work.com",
		"--token", "tok123",
		"--user", "alice",
		"--auth-user", "alice@work.com",
		"--skip-tls",
		"--backend", "server",
		"--git-protocol", "https",
	})
	require.NoError(t, cmd.Execute())

	p, ok := store.Get("work")
	require.True(t, ok)
	assert.Equal(t, "alice", p.User)
	assert.Equal(t, "alice@work.com", p.AuthUser)
	assert.True(t, p.SkipTLSVerify)
	assert.Equal(t, "server", p.BackendType)
	assert.Equal(t, "https", p.GitProtocol)
}

func TestCreateRun_RequiredFlagsMissing(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	// missing both --hostname and --token
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCreateRun_RequiredHostnameMissing(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work", "--token", "tok"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCreateRun_RequiredTokenMissing(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work", "--hostname", "git.work.com"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestCreateRun_InvalidBackend(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work", "--hostname", "git.work.com", "--token", "tok", "--backend", "clud"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid backend_type")
}

func TestCreateRun_ValidBackendValues(t *testing.T) {
	t.Parallel()
	for _, backend := range []string{"cloud", "server", "datacenter", ""} {
		backend := backend
		t.Run(backend, func(t *testing.T) {
			t.Parallel()
			f, _, _ := factorytest.New(t, factorytest.Opts{})
			store := profiles.New(t.TempDir())
			factorytest.UseProfiles(f, store)

			args := []string{"work", "--hostname", "git.work.com", "--token", "tok"}
			if backend != "" {
				args = append(args, "--backend", backend)
			}
			cmd := create.NewCmdCreate(f, nil)
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
		})
	}
}

func TestCreateRun_DuplicateNameOverwrites(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	// First create
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"work", "--hostname", "old.com", "--token", "old-tok"})
	require.NoError(t, cmd.Execute())

	// Overwrite
	cmd2 := create.NewCmdCreate(f, nil)
	cmd2.SetArgs([]string{"work", "--hostname", "new.com", "--token", "new-tok"})
	require.NoError(t, cmd2.Execute())

	p, ok := store.Get("work")
	require.True(t, ok)
	assert.Equal(t, "new.com", p.Hostname)
	assert.Equal(t, "new-tok", p.Token)
}
