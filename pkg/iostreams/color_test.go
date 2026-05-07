package iostreams_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func TestIOStreams_Color_DisabledReturnsPlain(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test() // colorEnabled = false
	assert.Equal(t, "SUCCESSFUL", ios.ColorGreen("SUCCESSFUL"))
	assert.Equal(t, "FAILED", ios.ColorRed("FAILED"))
	assert.Equal(t, "MERGED", ios.ColorMagenta("MERGED"))
}

func TestIOStreams_Color_EnabledWrapsANSI(t *testing.T) {
	t.Parallel()
	ios := iostreams.TestTTY() // colorEnabled = true
	assert.Equal(t, "\033[32mSUCCESSFUL\033[0m", ios.ColorGreen("SUCCESSFUL"))
	assert.Equal(t, "\033[31mFAILED\033[0m", ios.ColorRed("FAILED"))
	assert.Equal(t, "\033[35mMERGED\033[0m", ios.ColorMagenta("MERGED"))
	assert.Equal(t, "\033[33mIN_PROGRESS\033[0m", ios.ColorYellow("IN_PROGRESS"))
}

func TestIOStreams_ColorYellow_DisabledReturnsPlain(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	assert.Equal(t, "PENDING", ios.ColorYellow("PENDING"))
}

func TestIOStreams_SetColorEnabled_TogglesDecision(t *testing.T) {
	t.Parallel()
	ios := iostreams.TestTTY()
	assert.True(t, ios.ColorEnabled())
	assert.Equal(t, "\033[32mOK\033[0m", ios.ColorGreen("OK"))

	// --no-color path: force-disable.
	ios.SetColorEnabled(false)
	assert.False(t, ios.ColorEnabled())
	assert.Equal(t, "OK", ios.ColorGreen("OK"))

	// Re-enabling restores wrapping. The setter is symmetric.
	ios.SetColorEnabled(true)
	assert.True(t, ios.ColorEnabled())
	assert.Equal(t, "\033[32mOK\033[0m", ios.ColorGreen("OK"))
}

func TestIOStreams_SetColorEnabled_TrueOnNonTTY_ForcesEnable(t *testing.T) {
	// Edge case: a Test() IOStreams starts non-TTY / colorEnabled=false.
	// Explicit SetColorEnabled(true) must win — caller knows what they want.
	t.Parallel()
	ios := iostreams.Test()
	assert.False(t, ios.ColorEnabled())
	ios.SetColorEnabled(true)
	assert.True(t, ios.ColorEnabled())
	assert.Equal(t, "\033[31mERR\033[0m", ios.ColorRed("ERR"))
}
