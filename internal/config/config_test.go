package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/config"
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

func TestHostConfig_BackendType_LoadCloud(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeHostsFile(t, dir, "bb.example.com:\n  user: alice\n  git_protocol: https\n  backend_type: cloud\n")
	c := config.New(dir)
	require.NoError(t, c.Load())
	hc, ok := c.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, "cloud", hc.BackendType)
}

func TestHostConfig_BackendType_LoadServer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeHostsFile(t, dir, "bb.example.com:\n  user: alice\n  git_protocol: https\n  backend_type: server\n")
	c := config.New(dir)
	require.NoError(t, c.Load())
	hc, ok := c.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, "server", hc.BackendType)
}

func TestHostConfig_BackendType_LoadMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeHostsFile(t, dir, "bb.example.com:\n  user: alice\n  git_protocol: https\n")
	c := config.New(dir)
	require.NoError(t, c.Load())
	hc, ok := c.Get("bb.example.com")
	require.True(t, ok)
	assert.Empty(t, hc.BackendType)
}

func TestHostConfig_BackendType_RoundTrip_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("a.example.com", config.HostConfig{User: "a", GitProtocol: "https"})
	require.NoError(t, c.Save())

	c2 := config.New(dir)
	require.NoError(t, c2.Load())
	got, ok := c2.Get("a.example.com")
	require.True(t, ok)
	assert.Empty(t, got.BackendType)
}

func TestHostConfig_BackendType_RoundTrip_Cloud(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("a.example.com", config.HostConfig{
		User:        "a",
		GitProtocol: "https",
		BackendType: "cloud",
	})
	require.NoError(t, c.Save())

	c2 := config.New(dir)
	require.NoError(t, c2.Load())
	got, ok := c2.Get("a.example.com")
	require.True(t, ok)
	assert.Equal(t, "cloud", got.BackendType)
}

// TestMarshalYAML_stripsToken verifies that saving a config with a token
// does not write the token to the YAML file on disk.
func TestMarshalYAML_stripsToken(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := config.New(dir)
	c.Set("h.example.com", config.HostConfig{
		User:        "alice",
		OAuthToken:  "super-secret-token",
		GitProtocol: "https",
	})
	require.NoError(t, c.Save())

	raw, err := os.ReadFile(filepath.Join(dir, "hosts.yml"))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "super-secret-token", "token must not appear in hosts.yml")
	assert.NotContains(t, string(raw), "oauth_token", "oauth_token key must not appear in hosts.yml")
}

// TestLoad_warnsMigration verifies that loading a hosts.yml that contains an
// oauth_token prints a migration warning to stderr.
func TestLoad_warnsMigration(t *testing.T) {
	// Not parallel — redirects os.Stderr.
	dir := t.TempDir()
	writeHostsFile(t, dir, "h.example.com:\n  user: alice\n  git_protocol: https\n  oauth_token: tok\n")

	// Redirect os.Stderr so we can capture the warning.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	origStderr := os.Stderr
	os.Stderr = w

	c := config.New(dir)
	loadErr := c.Load()

	require.NoError(t, w.Close())
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	require.NoError(t, loadErr)
	assert.True(t, strings.Contains(buf.String(), "auth migrate"),
		"expected migration hint in stderr, got: %q", buf.String())
	assert.True(t, strings.Contains(buf.String(), "h.example.com"),
		"expected hostname in warning, got: %q", buf.String())
}
