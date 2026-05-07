package cmdutil

import (
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// ExplainError writes a user-actionable message to ios.ErrOut.
//
// This is now a thin shim over pkg/errfmt — kept only to avoid touching
// every caller in cmd/bitbottle. New presentation logic (codes, hints,
// TTY colour) lives in errfmt.Render; see that package for the full
// resolution order.
func ExplainError(ios *iostreams.IOStreams, err error) {
	errfmt.Render(ios, err)
}
