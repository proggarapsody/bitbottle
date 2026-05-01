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
}
