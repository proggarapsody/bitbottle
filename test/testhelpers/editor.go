package testhelpers

import "os"

// FakeEditorLauncher is a test double for an editor.Launcher. When Edit is
// called it records the filename, writes Content to the file, and returns Err.
type FakeEditorLauncher struct {
	Files   []string
	Content string
	Err     error
}

// Edit records the filename and (if Err is nil) writes Content into it.
func (e *FakeEditorLauncher) Edit(filename string) error {
	e.Files = append(e.Files, filename)
	if e.Err != nil {
		return e.Err
	}
	if filename == "" {
		return nil
	}
	return os.WriteFile(filename, []byte(e.Content), 0o600)
}
