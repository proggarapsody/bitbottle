package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeHostsFile(t *testing.T, dir, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hosts.yml"), []byte(content), 0o600))
}

func TestConfig_ReadMissingFile(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "does-not-exist")
	c := config.New(dir)
	err := c.Load()
	require.Error(t, err)
}

func TestConfig_ReadEmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeHostsFile(t, dir, "")
	c := config.New(dir)
	require.NoError(t, c.Load())
	assert.Len(t, c.Hosts(), 0)
}

func TestConfig_ReadMalformedYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeHostsFile(t, dir, "not: valid: yaml: :::")
	c := config.New(dir)
	require.Error(t, c.Load())
}

func TestConfig_GetMissingHost(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	_, ok := c.Get("unknown.example.com")
	assert.False(t, ok)
}

func TestConfig_SetAndGet(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	hc := config.HostConfig{
		User:        "alice",
		OAuthToken:  "tok",
		GitProtocol: "ssh",
	}
	c.Set("bb.example.com", hc)
	got, ok := c.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, hc, got)
}

func TestConfig_Remove(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	c.Set("h.example.com", config.HostConfig{User: "u", GitProtocol: "https"})
	c.Remove("h.example.com")
	_, ok := c.Get("h.example.com")
	assert.False(t, ok)
}

func TestConfig_Hosts(t *testing.T) {
	t.Parallel()

	c := config.New(t.TempDir())
	c.Set("a.example.com", config.HostConfig{User: "a", GitProtocol: "https"})
	c.Set("b.example.com", config.HostConfig{User: "b", GitProtocol: "ssh"})
	hosts := c.Hosts()
	assert.ElementsMatch(t, []string{"a.example.com", "b.example.com"}, hosts)
}

func TestConfig_AtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("a.example.com", config.HostConfig{User: "a", GitProtocol: "https"})
	require.NoError(t, c.Save())

	// hosts.yml should exist; no stray temp files should remain
	_, err := os.Stat(filepath.Join(dir, "hosts.yml"))
	require.NoError(t, err)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.Equal(t, "hosts.yml", e.Name(), "unexpected leftover file %q", e.Name())
	}
}

func TestConfig_MultiHost(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("a.example.com", config.HostConfig{User: "a", GitProtocol: "https"})
	c.Set("b.example.com", config.HostConfig{User: "b", GitProtocol: "ssh"})
	require.NoError(t, c.Save())

	c2 := config.New(dir)
	require.NoError(t, c2.Load())

	a, ok := c2.Get("a.example.com")
	require.True(t, ok)
	assert.Equal(t, "a", a.User)
	assert.Equal(t, "https", a.GitProtocol)

	b, ok := c2.Get("b.example.com")
	require.True(t, ok)
	assert.Equal(t, "b", b.User)
	assert.Equal(t, "ssh", b.GitProtocol)
}

func TestConfig_SkipTLSVerify(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("a.example.com", config.HostConfig{
		User:          "a",
		GitProtocol:   "https",
		SkipTLSVerify: true,
	})
	require.NoError(t, c.Save())

	c2 := config.New(dir)
	require.NoError(t, c2.Load())
	got, ok := c2.Get("a.example.com")
	require.True(t, ok)
	assert.True(t, got.SkipTLSVerify)
}
