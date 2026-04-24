package iostreams

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// IOStreams holds the three standard streams and TTY metadata.
// Tests supply bytes.Buffer for In/Out/ErrOut and set IsStdoutTTY to a
// deterministic value rather than probing the real file descriptor.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	IsStdoutTTY func() bool
	IsStderrTTY func() bool

	colorEnabled bool
}

// System returns an IOStreams backed by real os.Stdin/Stdout/Stderr with
// real isatty-based TTY detection.
func System() *IOStreams {
	stdoutIsTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	stderrIsTTY := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
	return &IOStreams{
		In:           os.Stdin,
		Out:          os.Stdout,
		ErrOut:       os.Stderr,
		IsStdoutTTY:  func() bool { return stdoutIsTTY },
		IsStderrTTY:  func() bool { return stderrIsTTY },
		colorEnabled: stdoutIsTTY && os.Getenv("NO_COLOR") == "",
	}
}

// Test returns a buffer-backed, non-TTY IOStreams suitable for unit tests.
func Test() *IOStreams {
	return &IOStreams{
		In:           io.NopCloser(strings.NewReader("")),
		Out:          &bytes.Buffer{},
		ErrOut:       &bytes.Buffer{},
		IsStdoutTTY:  func() bool { return false },
		IsStderrTTY:  func() bool { return false },
		colorEnabled: false,
	}
}

// TestTTY returns a buffer-backed IOStreams that reports IsStdoutTTY = true.
func TestTTY() *IOStreams {
	ios := Test()
	ios.IsStdoutTTY = func() bool { return true }
	ios.IsStderrTTY = func() bool { return true }
	ios.colorEnabled = true
	return ios
}

// ColorEnabled reports whether color output should be emitted.
func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled
}

// StartPager is a no-op stub.
func (s *IOStreams) StartPager() error { return nil }

// StopPager is a no-op stub.
func (s *IOStreams) StopPager() {}
