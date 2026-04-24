package testhelpers

// FakeBrowserLauncher is a test double for a browser.Launcher.
type FakeBrowserLauncher struct {
	URLs []string
	Err  error
}

// Browse records the URL and returns the configured error, if any.
func (b *FakeBrowserLauncher) Browse(url string) error {
	b.URLs = append(b.URLs, url)
	return b.Err
}
