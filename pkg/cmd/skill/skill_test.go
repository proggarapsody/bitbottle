package skill_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/skill"
)

// TestSkillPath_PrintsCanonicalRoot verifies the trivial case — `path`
// should always succeed and print the documented install root, since
// users may run it from a fresh shell with no Node/agent setup at all.
func TestSkillPath_PrintsCanonicalRoot(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{})
	cmd := skill.NewCmdSkill(f)
	cmd.SetArgs([]string{"path"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "~/.agents/skills/bitbottle",
		"skill path output must include the canonical root")
}

// TestSkillInstall_NoNpx_ReturnsActionableError verifies the failure
// mode when Node isn't installed: the error must explain WHAT to do
// next, not just that something broke. Most users on Homebrew or Go
// install won't have Node, and this is their entry point.
func TestSkillInstall_NoNpx_ReturnsActionableError(t *testing.T) {
	// Cannot t.Parallel because we mutate PATH globally.
	origPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })
	require.NoError(t, os.Setenv("PATH", "/nonexistent-dir-for-test"))

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := skill.NewCmdSkill(f)
	cmd.SetArgs([]string{"install"})
	err := cmd.Execute()

	require.Error(t, err)
	msg := strings.ToLower(err.Error())
	assert.Contains(t, msg, "npx",
		"error must name the missing tool so users know what to install")
	assert.Contains(t, msg, "node",
		"error must point users to Node.js as the source of npx")
}
