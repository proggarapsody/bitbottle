// Package extensions manages bitbottle CLI extensions: install, list.
// Extensions are stored under ~/.config/bitbottle/extensions/<name>/.
package extensions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Doer is the HTTP transport interface used by Manager. It is a subset of
// *http.Client so tests can inject an httptest.Server-backed stub.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Manifest is written to <extDir>/<name>/manifest.json after installation.
type Manifest struct {
	Name    string `json:"name"`
	Repo    string `json:"repo,omitempty"`    // "owner/repo", empty for local
	Version string `json:"version,omitempty"` // tag_name from GitHub, empty for local
	SHA256  string `json:"sha256,omitempty"`  // hex SHA-256 of the binary, empty for local
	Local   bool   `json:"local"`
	// LocalPath is the original source path for local installs.
	LocalPath string `json:"local_path,omitempty"`
}

// Extension represents an installed extension as returned by List.
type Extension struct {
	Manifest
	BinPath string // absolute path to the installed binary / symlink
}

// Manager handles extension lifecycle under a root directory.
type Manager struct {
	dir    string // root extensions dir, e.g. ~/.config/bitbottle/extensions
	client Doer
	runner func(bin string, args []string, env []string) error // nil = defaultRunner
}

// defaultRunner executes bin with args and env, inheriting stdio.
func defaultRunner(bin string, args []string, env []string) error {
	cmd := exec.CommandContext(context.Background(), bin, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// New returns a Manager rooted at dir using client for HTTP requests.
func New(dir string, client Doer) *Manager {
	if client == nil {
		client = &http.Client{}
	}
	return &Manager{dir: dir, client: client, runner: defaultRunner}
}

// WithRunner returns a copy of m with a custom runner (for testing).
func (m *Manager) WithRunner(r func(bin string, args []string, env []string) error) *Manager {
	cp := *m
	cp.runner = r
	return &cp
}

// extDir returns the directory for a single extension.
func (m *Manager) extDir(name string) string {
	return filepath.Join(m.dir, name)
}

// binPath returns the path of the installed binary for an extension.
func (m *Manager) binPath(name string) string {
	return filepath.Join(m.extDir(name), "bin", name)
}

// manifestPath returns the manifest file path for an extension.
func (m *Manager) manifestPath(name string) string {
	return filepath.Join(m.extDir(name), "manifest.json")
}

// isInstalled reports whether an extension with the given name is installed.
func (m *Manager) isInstalled(name string) bool {
	_, err := os.Stat(m.manifestPath(name))
	return err == nil
}

// readManifest reads and parses the manifest for an installed extension.
func (m *Manager) readManifest(name string) (Manifest, error) {
	data, err := os.ReadFile(m.manifestPath(name))
	if err != nil {
		return Manifest{}, err
	}
	var mf Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return Manifest{}, fmt.Errorf("corrupt manifest for %q: %w", name, err)
	}
	return mf, nil
}

// writeManifest serialises mf to disk.
func (m *Manager) writeManifest(name string, mf Manifest) error {
	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.manifestPath(name), data, 0o644)
}

// InstallFromGitHub downloads the latest binary release for owner/repo from
// GitHub and installs it. If force is false and the extension is already
// installed it returns an error.
func (m *Manager) InstallFromGitHub(ownerRepo string, force bool) error {
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository %q: expected USER/REPO", ownerRepo)
	}
	owner, repo := parts[0], parts[1]

	if !strings.HasPrefix(repo, "bitbottle-") {
		return fmt.Errorf("repository name must start with \"bitbottle-\" (got: %s)", repo)
	}
	name := strings.TrimPrefix(repo, "bitbottle-")

	if !force && m.isInstalled(name) {
		return fmt.Errorf("extension %s is already installed; use --force to reinstall", name)
	}

	rel, err := m.fetchLatestRelease(owner, repo)
	if err != nil {
		return fmt.Errorf("fetching release for %s: %w", ownerRepo, err)
	}

	asset, ok := pickAsset(rel.Assets)
	if !ok {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
	}

	data, err := m.downloadURL(asset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", asset.Name, err)
	}

	sum := sha256sum(data)

	binDir := filepath.Join(m.extDir(name), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	binDest := m.binPath(name)
	if err := os.WriteFile(binDest, data, 0o755); err != nil {
		return err
	}

	mf := Manifest{
		Name:    name,
		Repo:    ownerRepo,
		Version: rel.TagName,
		SHA256:  sum,
		Local:   false,
	}
	return m.writeManifest(name, mf)
}

// InstallLocal symlinks path as an extension. path must be a directory
// containing a binary named bitbottle-<name> (or its basename is used).
// The installed name is derived from the last element of path, stripping
// the "bitbottle-" prefix if present.
func (m *Manager) InstallLocal(path string, force bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	base := filepath.Base(abs)
	repoName := base
	if !strings.HasPrefix(repoName, "bitbottle-") {
		return fmt.Errorf("repository name must start with \"bitbottle-\" (got: %s)", repoName)
	}
	name := strings.TrimPrefix(repoName, "bitbottle-")

	if !force && m.isInstalled(name) {
		return fmt.Errorf("extension %s is already installed; use --force to reinstall", name)
	}

	// The binary inside the local directory should be named bitbottle-<name>.
	// We symlink to the directory's binary directly.
	srcBin := filepath.Join(abs, "bin", repoName)
	if _, err := os.Stat(srcBin); err != nil {
		// Fallback: top-level binary with the full name.
		srcBin = filepath.Join(abs, repoName)
		if _, err2 := os.Stat(srcBin); err2 != nil {
			return fmt.Errorf("no binary found at %s/bin/%s or %s/%s", abs, repoName, abs, repoName)
		}
	}

	binDir := filepath.Join(m.extDir(name), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	binDest := m.binPath(name)
	// Remove existing symlink/file before relinking.
	_ = os.Remove(binDest)
	if err := os.Symlink(srcBin, binDest); err != nil {
		return err
	}

	mf := Manifest{
		Name:      name,
		Local:     true,
		LocalPath: abs,
	}
	return m.writeManifest(name, mf)
}

// Remove deletes the named extension entirely.
func (m *Manager) Remove(name string) error {
	if !m.isInstalled(name) {
		return fmt.Errorf("extension %q is not installed", name)
	}
	return os.RemoveAll(m.extDir(name))
}

// Upgrade upgrades one named extension to the latest GitHub release.
// Returns (oldVersion, newVersion, nil) on success.
// If local: returns ("", "", nil) — skipped.
// If already at latest and !force: returns (v, v, nil).
func (m *Manager) Upgrade(name string, force bool) (oldVer, newVer string, err error) {
	if !m.isInstalled(name) {
		return "", "", fmt.Errorf("extension %q is not installed", name)
	}
	mf, err := m.readManifest(name)
	if err != nil {
		return "", "", err
	}
	if mf.Local {
		return "", "", nil
	}

	parts := strings.SplitN(mf.Repo, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo in manifest: %q", mf.Repo)
	}
	owner, repo := parts[0], parts[1]

	rel, err := m.fetchLatestRelease(owner, repo)
	if err != nil {
		return "", "", fmt.Errorf("fetching release for %s: %w", mf.Repo, err)
	}

	oldVer = mf.Version
	if rel.TagName == oldVer && !force {
		return oldVer, oldVer, nil
	}

	if err := m.InstallFromGitHub(mf.Repo, true); err != nil {
		return "", "", err
	}
	return oldVer, rel.TagName, nil
}

// UpgradeAll upgrades every non-local installed extension.
// Returns map[name]error — nil error means upgraded or already up to date.
func (m *Manager) UpgradeAll(force bool) map[string]error {
	exts, err := m.List()
	if err != nil {
		return map[string]error{"": err}
	}
	results := make(map[string]error, len(exts))
	for _, e := range exts {
		if e.Local {
			results[e.Name] = nil
			continue
		}
		_, _, upgradeErr := m.Upgrade(e.Name, force)
		results[e.Name] = upgradeErr
	}
	return results
}

// List returns all installed extensions in alphabetical order.
func (m *Manager) List() ([]Extension, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var exts []Extension
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		mf, err := m.readManifest(name)
		if err != nil {
			// Skip corrupted entries silently.
			continue
		}
		exts = append(exts, Extension{
			Manifest: mf,
			BinPath:  m.binPath(name),
		})
	}
	return exts, nil
}

// Exec runs the named extension binary with args.
// It strips env vars whose name (before "=") contains "KEYRING_PASSPHRASE" or
// "KEYRING_PASSWORD" (case-insensitive), injects BB_TOKEN and
// BITBOTTLE_VERSION, then delegates to the manager's runner.
//
// Exit-code propagation: callers should check whether the returned error is
// *exec.ExitError and call os.Exit(ee.ExitCode()) so the shell sees the
// extension's exit code.
func (m *Manager) Exec(name string, args []string, token, version string) error {
	if !m.isInstalled(name) {
		return fmt.Errorf("extension %q is not installed", name)
	}
	binPath := m.binPath(name)
	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("extension binary not found at %s: %w", binPath, err)
	}

	// Build env: filter secrets, then inject bitbottle vars.
	raw := os.Environ()
	env := make([]string, 0, len(raw)+2)
	for _, kv := range raw {
		key := kv
		if idx := strings.IndexByte(kv, '='); idx >= 0 {
			key = kv[:idx]
		}
		upper := strings.ToUpper(key)
		if strings.Contains(upper, "KEYRING_PASSPHRASE") || strings.Contains(upper, "KEYRING_PASSWORD") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, "BB_TOKEN="+token)
	env = append(env, "BITBOTTLE_VERSION="+version)

	runner := m.runner
	if runner == nil {
		runner = defaultRunner
	}
	return runner(binPath, args, env)
}

// --- GitHub release API types ---

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (m *Manager) fetchLatestRelease(owner, repo string) (ghRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	if err != nil {
		return ghRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := m.client.Do(req)
	if err != nil {
		return ghRelease{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ghRelease{}, fmt.Errorf("GitHub API returned status %d for %s/%s", resp.StatusCode, owner, repo)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return ghRelease{}, fmt.Errorf("decoding release JSON: %w", err)
	}
	return rel, nil
}

// pickAsset selects the asset whose name contains both GOOS and GOARCH
// (using common packaging conventions like linux_amd64, darwin_arm64).
func pickAsset(assets []ghAsset) (ghAsset, bool) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, goos) && strings.Contains(lower, goarch) {
			return a, true
		}
	}
	return ghAsset{}, false
}

func (m *Manager) downloadURL(rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func sha256sum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
