package alias_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/alias"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func TestAliasSet_PersistsToDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{ConfigDir: dir})
	cmd := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
	cmd.SetArgs([]string{"set", "prs", "pr list --author @me"})
	require.NoError(t, cmd.Execute())

	assert.FileExists(t, filepath.Join(dir, "aliases.yml"))
}

func TestAliasList_PrintsEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{ConfigDir: dir})

	for _, args := range [][]string{
		{"set", "prs", "pr list --author @me"},
		{"set", "co", "!git checkout"},
	} {
		c := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
		c.SetArgs(args)
		require.NoError(t, c.Execute())
	}
	out.Reset()

	c := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
	c.SetArgs([]string{"list"})
	require.NoError(t, c.Execute())
	got := out.String()
	assert.Contains(t, got, "prs:")
	assert.Contains(t, got, "co:")
	assert.Contains(t, got, "!git checkout")
}

func TestAliasDelete_RemovesEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{ConfigDir: dir})

	set := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
	set.SetArgs([]string{"set", "prs", "pr list"})
	require.NoError(t, set.Execute())

	del := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
	del.SetArgs([]string{"delete", "prs"})
	require.NoError(t, del.Execute())

	store, err := f.Aliases()
	require.NoError(t, err)
	_, ok := store.Get("prs")
	assert.False(t, ok)
}

func TestAliasSet_RejectsBuiltinShadow(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{ConfigDir: t.TempDir()})
	cmd := alias.NewCmdAlias(f, []string{"pr", "repo", "alias"})
	cmd.SetArgs([]string{"set", "pr", "something else"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shadow")
}
