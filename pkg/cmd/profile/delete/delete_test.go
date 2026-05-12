package delete_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/profile/delete"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func storeWithProfile(t *testing.T) *profiles.Store {
	t.Helper()
	store := profiles.New(t.TempDir())
	require.NoError(t, store.Load())
	store.Set("work", profiles.Profile{Hostname: "git.work.com", Token: "tok"})
	return store
}

func TestDeleteRun_ConfirmFlag(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := storeWithProfile(t)
	factorytest.UseProfiles(f, store)

	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"work", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Deleted profile work")
	_, ok := store.Get("work")
	assert.False(t, ok)
}

func TestDeleteRun_NonInteractiveWithoutConfirm(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := storeWithProfile(t)
	factorytest.UseProfiles(f, store)
	// factorytest sets IsStdoutTTY → false, which is non-interactive

	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"work"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required")

	// Profile should still exist
	_, ok := store.Get("work")
	assert.True(t, ok)
}

func TestDeleteRun_ProfileNotFound(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"nonexistent", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `profile "nonexistent" not found`)
}

func TestDeleteRun_TTYPromptYes(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := storeWithProfile(t)
	factorytest.UseProfiles(f, store)
	// Simulate TTY with "y" input.
	f.IOStreams = &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader("y\n")),
		Out:         out,
		ErrOut:      f.IOStreams.ErrOut,
		IsStdoutTTY: func() bool { return true },
		IsStderrTTY: func() bool { return false },
	}

	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"work"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Deleted profile work")
	_, ok := store.Get("work")
	assert.False(t, ok)
}

func TestDeleteRun_TTYPromptNo(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := storeWithProfile(t)
	factorytest.UseProfiles(f, store)
	// Simulate TTY with "n" input.
	f.IOStreams = &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader("n\n")),
		Out:         out,
		ErrOut:      f.IOStreams.ErrOut,
		IsStdoutTTY: func() bool { return true },
		IsStderrTTY: func() bool { return false },
	}

	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"work"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Aborted.")
	// Profile should still exist.
	_, ok := store.Get("work")
	assert.True(t, ok)
}
