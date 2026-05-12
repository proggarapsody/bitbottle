package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/profile/list"
)

func newStoreWith(t *testing.T, entries map[string]profiles.Profile) *profiles.Store {
	t.Helper()
	store := profiles.New(t.TempDir())
	require.NoError(t, store.Load())
	for name, p := range entries {
		store.Set(name, p)
	}
	return store
}

func TestListRun_Empty(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := profiles.New(t.TempDir())
	factorytest.UseProfiles(f, store)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	assert.Empty(t, out.String())
}

func TestListRun_ListsProfiles(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := newStoreWith(t, map[string]profiles.Profile{
		"work":     {Hostname: "git.work.com", Token: "tok-w", User: "alice"},
		"personal": {Hostname: "bitbucket.org", Token: "tok-p"},
	})
	factorytest.UseProfiles(f, store)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "personal")
	assert.Contains(t, output, "bitbucket.org")
	assert.Contains(t, output, "work")
	assert.Contains(t, output, "git.work.com")
	assert.Contains(t, output, "alice")
	// Token must never appear in output.
	assert.NotContains(t, output, "tok-w")
	assert.NotContains(t, output, "tok-p")
}

func TestListRun_JSONMode(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := newStoreWith(t, map[string]profiles.Profile{
		"work": {Hostname: "git.work.com", Token: "secret", User: "alice", BackendType: "server"},
	})
	factorytest.UseProfiles(f, store)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"--json", "name,hostname,backend_type,skip_tls_verify"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, `"name"`)
	assert.Contains(t, output, `"work"`)
	assert.Contains(t, output, `"hostname"`)
	assert.Contains(t, output, `"git.work.com"`)
	assert.Contains(t, output, `"backend_type"`)
	assert.Contains(t, output, `"server"`)
	// Token must never appear in JSON output.
	assert.NotContains(t, output, "secret")
}

func TestListRun_TokenNeverPrinted(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := newStoreWith(t, map[string]profiles.Profile{
		"work": {Hostname: "git.work.com", Token: "ultra-secret-token"},
	})
	factorytest.UseProfiles(f, store)

	// Plain table
	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())
	assert.NotContains(t, out.String(), "ultra-secret-token")
}

func TestListRun_SortedOutput(t *testing.T) {
	f, out, _ := factorytest.New(t, factorytest.Opts{})
	store := newStoreWith(t, map[string]profiles.Profile{
		"zebra": {Hostname: "z.com", Token: "t"},
		"alpha": {Hostname: "a.com", Token: "t"},
	})
	factorytest.UseProfiles(f, store)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	output := out.String()
	// alpha should appear before zebra in the output
	aPos := len(output) - len(output[indexStr(output, "alpha"):])
	zPos := len(output) - len(output[indexStr(output, "zebra"):])
	assert.Less(t, aPos, zPos, "alpha should appear before zebra")
}

// indexStr returns the byte index of substr in s, or len(s) if not found.
func indexStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return len(s)
}
