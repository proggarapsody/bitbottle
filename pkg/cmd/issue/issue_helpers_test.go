package issue_test

import (
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// testTTYIOStreams is a thin alias around iostreams.TestTTY so each test
// file's intent reads as "I want a TTY-like IOStreams" rather than reaching
// into the iostreams package directly.
func testTTYIOStreams() *iostreams.IOStreams { return iostreams.TestTTY() }
