package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func FatalIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type BrowserLauncher interface {
	Browse(url string) error
}

type EditorLauncher interface {
	Edit(filename string) error
}

type SystemBrowser struct{}

// Browse opens the given URL in the system browser.
func (b *SystemBrowser) Browse(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url) //nolint:gosec,noctx
	default:
		cmd = exec.Command("open", url) //nolint:gosec,noctx
	}
	return cmd.Run()
}

type SystemEditor struct{}

// Edit opens the given filename in $EDITOR.
func (e *SystemEditor) Edit(filename string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, filename) //nolint:gosec,noctx
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
