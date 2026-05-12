package keyring_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/keyring"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestFakeKeyring_GetMissing(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	_, err := kr.Get("svc", "user")
	require.Error(t, err)
}

func TestFakeKeyring_SetAndGet(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("svc", "user", "secret"))
	got, err := kr.Get("svc", "user")
	require.NoError(t, err)
	assert.Equal(t, "secret", got)
}

func TestFakeKeyring_Delete(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("svc", "user", "secret"))
	require.NoError(t, kr.Delete("svc", "user"))
	_, err := kr.Get("svc", "user")
	require.Error(t, err)
}

func TestFakeKeyring_GetErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom get")
	kr := testhelpers.NewFakeKeyring()
	kr.GetErr = injected
	_, err := kr.Get("svc", "user")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_SetErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom set")
	kr := testhelpers.NewFakeKeyring()
	kr.SetErr = injected
	err := kr.Set("svc", "user", "pw")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_DeleteErr(t *testing.T) {
	t.Parallel()

	injected := errors.New("boom del")
	kr := testhelpers.NewFakeKeyring()
	kr.DelErr = injected
	err := kr.Delete("svc", "user")
	require.ErrorIs(t, err, injected)
}

func TestFakeKeyring_Concurrent(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			user := fmt.Sprintf("u%d", i)
			_ = kr.Set("svc", user, fmt.Sprintf("pw%d", i))
			_, _ = kr.Get("svc", user)
			_ = kr.Delete("svc", user)
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// IsHeadless tests
// ---------------------------------------------------------------------------

// TestIsHeadless_CI verifies that CI=1 causes IsHeadless to return true.
func TestIsHeadless_CI(t *testing.T) {
	t.Setenv("CI", "1")
	// Clear vars that might short-circuit the check before CI is evaluated on
	// some platforms.
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("DOCKER", "")
	assert.True(t, keyring.IsHeadless())
}

// TestIsHeadless_Interactive verifies that when all headless indicators are
// absent, IsHeadless returns false. We simulate an interactive desktop session
// by ensuring CI/SSH_TTY are unset and DISPLAY is set.
func TestIsHeadless_Interactive(t *testing.T) {
	// Vars that must be absent (not just empty) to avoid false positives.
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "DOCKER", "SSH_TTY", "WAYLAND_DISPLAY"} {
		old, had := os.LookupEnv(v)
		if err := os.Unsetenv(v); err != nil {
			t.Fatalf("Unsetenv %s: %v", v, err)
		}
		if had {
			t.Cleanup(func() { os.Setenv(v, old) }) //nolint:errcheck
		} else {
			t.Cleanup(func() { os.Unsetenv(v) }) //nolint:errcheck
		}
	}
	// Simulate an active DBus session so Linux does not treat this as headless.
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "/run/user/1000/bus")
	// Set DISPLAY to simulate an interactive session.
	t.Setenv("DISPLAY", ":0")
	assert.False(t, keyring.IsHeadless())
}

// ---------------------------------------------------------------------------
// New() constructor tests
// ---------------------------------------------------------------------------

// TestNew_returnsOSKeyring verifies that New returns *OSKeyring by default.
func TestNew_returnsOSKeyring(t *testing.T) {
	t.Setenv("BITBOTTLE_ALLOW_INSECURE_STORE", "")
	kr := keyring.New()
	_, ok := kr.(*keyring.OSKeyring)
	assert.True(t, ok, "expected *keyring.OSKeyring, got %T", kr)
}

// TestNew_returnsFileKeyring verifies that BITBOTTLE_ALLOW_INSECURE_STORE=1
// causes New to return *FileKeyring.
func TestNew_returnsFileKeyring(t *testing.T) {
	t.Setenv("BITBOTTLE_ALLOW_INSECURE_STORE", "1")
	kr := keyring.New()
	_, ok := kr.(*keyring.FileKeyring)
	assert.True(t, ok, "expected *keyring.FileKeyring, got %T", kr)
}

// ---------------------------------------------------------------------------
// FileKeyring round-trip
// ---------------------------------------------------------------------------

// TestFileKeyring_RoundTrip verifies set → get → delete using the file store.
func TestFileKeyring_RoundTrip(t *testing.T) {
	// t.Setenv and t.Parallel are mutually exclusive; no t.Parallel here.
	t.Setenv("BITBOTTLE_ALLOW_INSECURE_STORE", "1")

	// Use a temp home so the test doesn't touch the real config dir.
	t.Setenv("HOME", t.TempDir())

	kr := &keyring.FileKeyring{}
	require.NoError(t, kr.Set("svc", "testuser", "hunter2"))

	got, err := kr.Get("svc", "testuser")
	require.NoError(t, err)
	assert.Equal(t, "hunter2", got)

	require.NoError(t, kr.Delete("svc", "testuser"))

	_, err = kr.Get("svc", "testuser")
	require.ErrorIs(t, err, keyring.ErrNotFound)
}

// TestFileKeyring_GetMissing verifies ErrNotFound when no file exists.
func TestFileKeyring_GetMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	kr := &keyring.FileKeyring{}
	_, err := kr.Get("svc", "nobody")
	require.ErrorIs(t, err, keyring.ErrNotFound)
}

// TestFileKeyring_DeleteIdempotent verifies that deleting a nonexistent key
// returns nil (idempotent).
func TestFileKeyring_DeleteIdempotent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	kr := &keyring.FileKeyring{}
	require.NoError(t, kr.Delete("svc", "ghost"))
}
