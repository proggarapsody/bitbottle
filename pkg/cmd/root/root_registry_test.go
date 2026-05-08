package root_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"

	// Side-effect import: registers context via init().
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/context"
)

// TestRoot_ContextRegisteredViaCmdRegistry verifies that the `context`
// subcommand is reachable from the root command after the pkg/cmd/context
// package self-registers through pkg/cmdregistry.
func TestRoot_ContextRegisteredViaCmdRegistry(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	rootCmd := root.NewCmdRoot(f)

	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "context" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'context' command to be present in root via cmdregistry")
}
