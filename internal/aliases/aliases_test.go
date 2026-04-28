package aliases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/aliases"
)

func TestResolve_NotFound_ReturnsFalse(t *testing.T) {
	t.Parallel()

	store := aliases.New(t.TempDir())
	exp, ok, err := aliases.Resolve(store, "nope", []string{"x"})
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, exp)
}

func TestResolve_CommandAlias_AppendsTrailingArgs(t *testing.T) {
	t.Parallel()

	store := aliases.New(t.TempDir())
	require.NoError(t, store.Set("prs", "pr list --author @me"))

	exp, ok, err := aliases.Resolve(store, "prs", []string{"--limit", "5"})
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, []string{"pr", "list", "--author", "@me", "--limit", "5"}, exp.Args)
	assert.Empty(t, exp.Shell)
}

func TestResolve_ShellAlias_InterpolatesPositional(t *testing.T) {
	t.Parallel()

	store := aliases.New(t.TempDir())
	require.NoError(t, store.Set("dep", "!ssh prod tail -f /var/log/$1.log"))

	exp, ok, err := aliases.Resolve(store, "dep", []string{"api"})
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "ssh prod tail -f /var/log/api.log", exp.Shell)
}

func TestResolve_ShellAlias_DollarAtJoinsAll(t *testing.T) {
	t.Parallel()

	store := aliases.New(t.TempDir())
	require.NoError(t, store.Set("echoall", "!echo $@"))

	exp, ok, err := aliases.Resolve(store, "echoall", []string{"a", "b c", "d"})
	require.NoError(t, err)
	require.True(t, ok)
	// shellquote.Join quotes "b c" because of the space.
	assert.Equal(t, "echo a 'b c' d", exp.Shell)
}

func TestSaveLoad_RoundTrips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s1 := aliases.New(dir)
	require.NoError(t, s1.Set("prs", "pr list"))
	require.NoError(t, s1.Save())

	s2 := aliases.New(dir)
	require.NoError(t, s2.Load())
	v, ok := s2.Get("prs")
	require.True(t, ok)
	assert.Equal(t, "pr list", v)
}
