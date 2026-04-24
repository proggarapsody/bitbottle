package testhelpers

import (
	"bytes"
	"io"
	"strings"
)

// IOStreams is a test double for pkg/iostreams.IOStreams. It will be replaced
// with the real import once pkg/iostreams exists.
type IOStreams struct {
	In           io.ReadCloser
	Out          io.Writer
	ErrOut       io.Writer
	IsStdoutTTY  func() bool
	IsStderrTTY  func() bool
	ColorEnabled bool
}

// nopReadCloser wraps an io.Reader to satisfy io.ReadCloser.
type nopReadCloser struct{ io.Reader }

func (nopReadCloser) Close() error { return nil }

// TestIOStreams returns an IOStreams with buffered stdout/stderr and an empty
// stdin, with both TTY flags reporting false.
func TestIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ios := &IOStreams{
		In:           nopReadCloser{strings.NewReader("")},
		Out:          stdout,
		ErrOut:       stderr,
		IsStdoutTTY:  func() bool { return false },
		IsStderrTTY:  func() bool { return false },
		ColorEnabled: false,
	}
	return ios, stdout, stderr
}

// TTYIOStreams returns an IOStreams that claims stdout and stderr are TTYs and
// has color enabled.
func TTYIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
	ios, stdout, stderr := TestIOStreams()
	ios.IsStdoutTTY = func() bool { return true }
	ios.IsStderrTTY = func() bool { return true }
	ios.ColorEnabled = true
	return ios, stdout, stderr
}
