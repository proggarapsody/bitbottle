package iostreams

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
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

	// pager fields
	pagerCmd *exec.Cmd
	pagerIn  io.WriteCloser
	pagerOut io.Writer // original Out, restored by StopPager
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

// StartPager spawns $PAGER (default "less -FRX") and wires IOStreams.Out
// to the pager's stdin. Only activates when stdout is a TTY; no-op otherwise.
// Callers must defer StopPager() immediately after a successful call.
func (s *IOStreams) StartPager() error {
	if !s.IsStdoutTTY() {
		return nil
	}

	pagerCmd := os.Getenv("PAGER")
	if pagerCmd == "" {
		pagerCmd = "less -FRX"
	}

	// context.Background() is intentional: the pager's lifetime is bounded by
	// the explicit StopPager call (tied to command completion via defer),
	// not by request cancellation.
	cmd := exec.CommandContext(context.Background(), "sh", "-c", pagerCmd)
	cmd.Stdout = s.Out
	cmd.Stderr = s.ErrOut

	pagerIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		_ = pagerIn.Close()
		// Fall through: write directly to Out without a pager.
		return nil
	}

	s.pagerOut = s.Out
	s.Out = pagerIn
	s.pagerIn = pagerIn
	s.pagerCmd = cmd
	return nil
}

// StopPager closes the pager's stdin and waits for the process to exit.
// Safe to call even when no pager was started.
func (s *IOStreams) StopPager() {
	if s.pagerCmd == nil {
		return
	}
	_ = s.pagerIn.Close()
	_ = s.pagerCmd.Wait()
	s.Out = s.pagerOut
	s.pagerCmd = nil
	s.pagerIn = nil
	s.pagerOut = nil
}
