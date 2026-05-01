package iostreams_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func TestIOStreams_StartPager_NonTTY_NoOp(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test() // non-TTY
	originalOut := ios.Out
	require.NoError(t, ios.StartPager())
	assert.Equal(t, originalOut, ios.Out, "Out should be unchanged when not a TTY")
	ios.StopPager() // should be safe no-op
}

func TestIOStreams_StopPager_WithoutStart_IsNoOp(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	ios.StopPager() // must not panic
}

func TestIOStreams_StartPager_TTY_WiresPagerProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns subprocess")
	}
	t.Setenv("PAGER", "cat")

	var buf bytes.Buffer
	ios := iostreams.TestTTY()
	ios.Out = &buf

	require.NoError(t, ios.StartPager())
	// After StartPager, ios.Out should be a pipe (not our original buffer).
	assert.NotEqual(t, &buf, ios.Out, "Out should be replaced by pager pipe")

	_, _ = fmt.Fprint(ios.Out, "hello pager")
	ios.StopPager()
	// After StopPager, output has gone through cat to our buffer.
	assert.Contains(t, buf.String(), "hello pager")
}
