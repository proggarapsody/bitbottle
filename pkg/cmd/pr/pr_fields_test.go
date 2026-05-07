package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// TestPRStateColor_TTY exercises every documented PR state plus an unknown
// one. We assert against the raw ANSI escape because the same string is what
// lands on the user's terminal — a stricter test than checking "ColorGreen
// was called" via a fake.
func TestPRStateColor_TTY(t *testing.T) {
	t.Parallel()
	colorize := pr.PRStateColor(iostreams.TestTTY())

	cases := map[string]string{
		"OPEN":       "\033[32mOPEN\033[0m",
		"MERGED":     "\033[35mMERGED\033[0m",
		"DECLINED":   "\033[31mDECLINED\033[0m",
		"SUPERSEDED": "\033[31mSUPERSEDED\033[0m",
		// Edge cases:
		"":        "",        // empty state — passes through
		"WEIRDO":  "WEIRDO",  // unknown state — no false-positive color
		"open":    "open",    // case-sensitive: lowercase ≠ OPEN
		"MERGED ": "MERGED ", // trailing space defeats exact match
	}
	for state, want := range cases {
		assert.Equal(t, want, colorize(state), "state=%q", state)
	}
}

// TestPRStateColor_NonTTY proves a piped invocation never emits ANSI bytes —
// otherwise downstream tools would see junk in the STATE column.
func TestPRStateColor_NonTTY(t *testing.T) {
	t.Parallel()
	colorize := pr.PRStateColor(iostreams.Test())
	for _, state := range []string{"OPEN", "MERGED", "DECLINED", "SUPERSEDED"} {
		assert.Equal(t, state, colorize(state), "non-TTY must not wrap %q in ANSI", state)
	}
}
