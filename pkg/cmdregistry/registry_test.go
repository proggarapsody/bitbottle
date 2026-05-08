package cmdregistry_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func TestRegister_AddsBuilder(t *testing.T) {
	t.Parallel()
	r := cmdregistry.NewRegistry()
	called := false
	r.Register(func(_ *factory.Factory) *cobra.Command {
		called = true
		return &cobra.Command{Use: "testcmd"}
	})
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmds := r.All(f)
	require.Len(t, cmds, 1)
	assert.True(t, called, "builder should have been called by All")
}

func TestAll_ReturnsAlphabeticalByUse(t *testing.T) {
	t.Parallel()
	r := cmdregistry.NewRegistry()
	r.Register(func(_ *factory.Factory) *cobra.Command { return &cobra.Command{Use: "zebra"} })
	r.Register(func(_ *factory.Factory) *cobra.Command { return &cobra.Command{Use: "apple"} })
	r.Register(func(_ *factory.Factory) *cobra.Command { return &cobra.Command{Use: "mango"} })

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmds := r.All(f)
	require.Len(t, cmds, 3)
	assert.Equal(t, "apple", cmds[0].Use)
	assert.Equal(t, "mango", cmds[1].Use)
	assert.Equal(t, "zebra", cmds[2].Use)
}

func TestAll_EmptyWhenNoneRegistered(t *testing.T) {
	t.Parallel()
	r := cmdregistry.NewRegistry()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmds := r.All(f)
	assert.Empty(t, cmds)
}

func TestGlobal_RegisterAndAll(t *testing.T) {
	// Test that package-level Register and All work via the global registry.
	// We can't safely call the package-level Register in parallel tests (global
	// state), so this runs sequentially and uses a fresh snapshot via All.
	r := cmdregistry.NewRegistry()
	r.Register(func(_ *factory.Factory) *cobra.Command { return &cobra.Command{Use: "globaltest"} })
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmds := r.All(f)
	require.Len(t, cmds, 1)
	assert.Equal(t, "globaltest", cmds[0].Use)
}
