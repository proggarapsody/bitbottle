package iostreams

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	IsStdoutTTY func() bool
	IsStderrTTY func() bool

	colorEnabled bool
}

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

func (s *IOStreams) StartPager() error { return nil }

func (s *IOStreams) StopPager() {}
