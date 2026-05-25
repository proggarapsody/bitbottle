package prsettings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	prsettings "github.com/proggarapsody/bitbottle/pkg/cmd/repo/pr-settings"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

// TestNewCmdPRSettings_HasSubcommands verifies that the group command registers
// both "get" and "set" subcommands.
func TestNewCmdPRSettings_HasSubcommands(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: testConfig})
	cmd := prsettings.NewCmdPRSettings(f)

	names := make([]string, 0)
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}

	assert.Contains(t, names, "get")
	assert.Contains(t, names, "set")
}
