package profiles_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/profiles"
)

func newStore(t *testing.T) (*profiles.Store, string) {
	t.Helper()
	dir := t.TempDir()
	return profiles.New(dir), dir
}

func TestLoad_MissingFile(t *testing.T) {
	s, _ := newStore(t)
	require.NoError(t, s.Load())
	assert.Empty(t, s.All())
}

func TestLoad_EmptyFile(t *testing.T) {
	s, dir := newStore(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "profiles.yml"), []byte(""), 0o600))
	require.NoError(t, s.Load())
	assert.Empty(t, s.All())
}

func TestSetGetDelete(t *testing.T) {
	s, _ := newStore(t)
	require.NoError(t, s.Load())

	p := profiles.Profile{
		Hostname: "git.example.com",
		Token:    "secret-token",
		User:     "alice",
	}
	s.Set("work", p)

	got, ok := s.Get("work")
	require.True(t, ok)
	assert.Equal(t, p, got)

	_, ok = s.Get("nonexistent")
	assert.False(t, ok)

	deleted := s.Delete("work")
	assert.True(t, deleted)

	_, ok = s.Get("work")
	assert.False(t, ok)

	deletedAgain := s.Delete("work")
	assert.False(t, deletedAgain)
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	s, dir := newStore(t)
	require.NoError(t, s.Load())

	s.Set("work", profiles.Profile{
		Hostname:      "git.work.com",
		Token:         "work-token",
		User:          "alice",
		AuthUser:      "alice@work.com",
		SkipTLSVerify: true,
		BackendType:   "server",
		GitProtocol:   "https",
	})
	s.Set("personal", profiles.Profile{
		Hostname:    "bitbucket.org",
		Token:       "bb-token",
		BackendType: "cloud",
		GitProtocol: "ssh",
	})

	require.NoError(t, s.Save())

	// Verify file was created.
	_, err := os.Stat(filepath.Join(dir, "profiles.yml"))
	require.NoError(t, err)

	// Load into a fresh store and verify round-trip.
	s2 := profiles.New(dir)
	require.NoError(t, s2.Load())

	work, ok := s2.Get("work")
	require.True(t, ok)
	assert.Equal(t, "git.work.com", work.Hostname)
	assert.Equal(t, "work-token", work.Token)
	assert.Equal(t, "alice", work.User)
	assert.Equal(t, "alice@work.com", work.AuthUser)
	assert.True(t, work.SkipTLSVerify)
	assert.Equal(t, "server", work.BackendType)
	assert.Equal(t, "https", work.GitProtocol)

	personal, ok := s2.Get("personal")
	require.True(t, ok)
	assert.Equal(t, "bitbucket.org", personal.Hostname)
	assert.Equal(t, "bb-token", personal.Token)
	assert.Equal(t, "cloud", personal.BackendType)
	assert.Equal(t, "ssh", personal.GitProtocol)
}

func TestAll_SortedByName(t *testing.T) {
	s, _ := newStore(t)
	require.NoError(t, s.Load())

	s.Set("zebra", profiles.Profile{Hostname: "z.com", Token: "t"})
	s.Set("alpha", profiles.Profile{Hostname: "a.com", Token: "t"})
	s.Set("middle", profiles.Profile{Hostname: "m.com", Token: "t"})

	all := s.All()
	require.Len(t, all, 3)
	assert.Equal(t, "alpha", all[0].Name)
	assert.Equal(t, "middle", all[1].Name)
	assert.Equal(t, "zebra", all[2].Name)
}

func TestAll_Empty(t *testing.T) {
	s, _ := newStore(t)
	require.NoError(t, s.Load())
	assert.Empty(t, s.All())
}

func TestSet_Overwrite(t *testing.T) {
	s, _ := newStore(t)
	require.NoError(t, s.Load())

	s.Set("work", profiles.Profile{Hostname: "old.com", Token: "old-token"})
	s.Set("work", profiles.Profile{Hostname: "new.com", Token: "new-token"})

	got, ok := s.Get("work")
	require.True(t, ok)
	assert.Equal(t, "new.com", got.Hostname)
	assert.Equal(t, "new-token", got.Token)
}

func TestSave_AtomicWriteCreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "bitbottle")
	s := profiles.New(dir)
	// no Load() needed — empty store
	s.Set("p", profiles.Profile{Hostname: "h.com", Token: "t"})
	require.NoError(t, s.Save())

	_, err := os.Stat(filepath.Join(dir, "profiles.yml"))
	require.NoError(t, err)
}
