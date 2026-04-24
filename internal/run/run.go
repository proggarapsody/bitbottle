package run

import (
	"bytes"
	"os"
	"os/exec"
)

// Runner abstracts shelling out to the git binary.
type Runner interface {
	Run(args ...string) (stdout, stderr string, err error)
	RunInteractive(args ...string) error
}

// SystemRunner executes real git commands.
type SystemRunner struct{}

func (r *SystemRunner) Run(args ...string) (string, string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func (r *SystemRunner) RunInteractive(args ...string) error {
	cmd := exec.Command("git", args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
