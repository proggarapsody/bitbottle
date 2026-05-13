package extensions_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/internal/extensions"
)

// makeRelease builds a minimal GitHub release JSON response with one asset
// whose name encodes goos/goarch so pickAsset can match it.
func makeRelease(tag, assetURL string) []byte {
	rel := map[string]any{
		"tag_name": tag,
		"assets": []map[string]any{
			{
				"name":                 fmt.Sprintf("bitbottle-hello_%s_%s", runtime.GOOS, runtime.GOARCH),
				"browser_download_url": assetURL,
			},
		},
	}
	data, _ := json.Marshal(rel)
	return data
}

// newTestServer returns an httptest.Server that serves the release JSON at
// /repos/owner/repo/releases/latest and the binary at /download.
func newTestServer(t *testing.T, releaseTag string, binaryContent []byte) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	downloadURL := srv.URL + "/download"
	relJSON := makeRelease(releaseTag, downloadURL)

	mux.HandleFunc("/repos/owner/bitbottle-hello/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(relJSON)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(binaryContent)
	})
	return srv
}

// newManagerWithDir creates a Manager rooted at extRoot using a rewrite
// transport so that calls to api.github.com land on srv.
func newManagerWithDir(extRoot string, srv *httptest.Server) *extensions.Manager {
	if srv == nil {
		return extensions.New(extRoot, nil)
	}
	rt := &rewriteTransport{base: srv.URL, inner: srv.Client().Transport}
	return extensions.New(extRoot, &http.Client{Transport: rt})
}

// TestInstallFromGitHub_HappyPath verifies that InstallFromGitHub downloads
// the binary, writes it under <extDir>/hello/bin/hello, and writes a manifest.
func TestInstallFromGitHub_HappyPath(t *testing.T) {
	binContent := []byte("#!/bin/sh\necho hello\n")
	srv := newTestServer(t, "v1.0.0", binContent)
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("InstallFromGitHub: %v", err)
	}

	// Binary must exist and be executable.
	binPath := filepath.Join(dir, "hello", "bin", "hello")
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("binary not found at %s: %v", binPath, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("binary is not executable: mode=%v", info.Mode())
	}

	// Manifest must exist with correct fields.
	mfPath := filepath.Join(dir, "hello", "manifest.json")
	data, err := os.ReadFile(mfPath)
	if err != nil {
		t.Fatalf("manifest not found: %v", err)
	}
	var mf map[string]any
	if err := json.Unmarshal(data, &mf); err != nil {
		t.Fatalf("corrupt manifest: %v", err)
	}
	if mf["version"] != "v1.0.0" {
		t.Errorf("manifest version = %v; want v1.0.0", mf["version"])
	}
	if mf["repo"] != "owner/bitbottle-hello" {
		t.Errorf("manifest repo = %v; want owner/bitbottle-hello", mf["repo"])
	}
}

// TestInstallFromGitHub_AlreadyInstalled checks that a second install without
// --force returns an error.
func TestInstallFromGitHub_AlreadyInstalled(t *testing.T) {
	srv := newTestServer(t, "v1.0.0", []byte("binary"))
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err == nil {
		t.Fatal("expected error on second install without --force; got nil")
	}
}

// TestInstallFromGitHub_MissingPrefix checks that repos without "bitbottle-"
// prefix are rejected.
func TestInstallFromGitHub_MissingPrefix(t *testing.T) {
	mgr := extensions.New(t.TempDir(), nil)
	if err := mgr.InstallFromGitHub("owner/hello", false); err == nil {
		t.Fatal("expected error for missing bitbottle- prefix; got nil")
	}
}

// TestInstallFromGitHub_NoMatchingAsset checks that a release with no
// GOOS/GOARCH-matching asset returns an error.
func TestInstallFromGitHub_NoMatchingAsset(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Release with an asset that never matches any platform.
	relJSON, _ := json.Marshal(map[string]any{
		"tag_name": "v1.0.0",
		"assets": []map[string]any{
			{
				"name":                 "bitbottle-hello_plan9_riscv64",
				"browser_download_url": srv.URL + "/download",
			},
		},
	})
	mux.HandleFunc("/repos/owner/bitbottle-hello/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(relJSON)
	})

	rt := &rewriteTransport{base: srv.URL, inner: srv.Client().Transport}
	mgr := extensions.New(t.TempDir(), &http.Client{Transport: rt})

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err == nil {
		t.Fatal("expected error for no matching asset; got nil")
	}
}

// TestInstallLocal_HappyPath verifies that a local directory with a binary is
// symlinked correctly and a manifest is written.
func TestInstallLocal_HappyPath(t *testing.T) {
	// Create a fake extension source directory.
	srcBase := t.TempDir()
	extName := "bitbottle-greet"
	srcDir := filepath.Join(srcBase, extName)
	if err := os.MkdirAll(filepath.Join(srcDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	srcBin := filepath.Join(srcDir, "bin", extName)
	if err := os.WriteFile(srcBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	extRoot := t.TempDir()
	mgr := extensions.New(extRoot, nil)
	if err := mgr.InstallLocal(srcDir, false); err != nil {
		t.Fatalf("InstallLocal: %v", err)
	}

	// Manifest must record Local=true.
	mfPath := filepath.Join(extRoot, "greet", "manifest.json")
	data, err := os.ReadFile(mfPath)
	if err != nil {
		t.Fatalf("manifest not found: %v", err)
	}
	var mf map[string]any
	if err := json.Unmarshal(data, &mf); err != nil {
		t.Fatalf("corrupt manifest: %v", err)
	}
	if mf["local"] != true {
		t.Errorf("manifest local = %v; want true", mf["local"])
	}
}

// TestInstallLocal_AlreadyInstalled checks the --force=false guard.
func TestInstallLocal_AlreadyInstalled(t *testing.T) {
	srcBase := t.TempDir()
	extName := "bitbottle-greet"
	srcDir := filepath.Join(srcBase, extName)
	if err := os.MkdirAll(filepath.Join(srcDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	srcBin := filepath.Join(srcDir, "bin", extName)
	if err := os.WriteFile(srcBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	extRoot := t.TempDir()
	mgr := extensions.New(extRoot, nil)
	if err := mgr.InstallLocal(srcDir, false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := mgr.InstallLocal(srcDir, false); err == nil {
		t.Fatal("expected error on second install without --force; got nil")
	}
}

// TestInstallLocal_NonBitbottlePrefix verifies rejection of dirs not starting
// with "bitbottle-".
func TestInstallLocal_NonBitbottlePrefix(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "myplugin")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mgr := extensions.New(t.TempDir(), nil)
	if err := mgr.InstallLocal(srcDir, false); err == nil {
		t.Fatal("expected error for non-bitbottle- prefix; got nil")
	}
}

// TestList_Empty verifies that List returns nil for a missing directory.
func TestList_Empty(t *testing.T) {
	mgr := extensions.New(filepath.Join(t.TempDir(), "nonexistent"), nil)
	exts, err := mgr.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if exts != nil {
		t.Errorf("List = %v; want nil", exts)
	}
}

// TestList_OneRemoteOneLocal verifies that List returns both a remote and a
// local extension from a pre-populated directory.
func TestList_OneRemoteOneLocal(t *testing.T) {
	extRoot := t.TempDir()

	// Remote extension.
	remoteName := "hello"
	remoteDir := filepath.Join(extRoot, remoteName)
	if err := os.MkdirAll(filepath.Join(remoteDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	remoteMF := map[string]any{
		"name":    remoteName,
		"repo":    "owner/bitbottle-hello",
		"version": "v1.2.3",
		"sha256":  "abc",
		"local":   false,
	}
	writeMF(t, filepath.Join(remoteDir, "manifest.json"), remoteMF)

	// Local extension.
	localName := "greet"
	localDir := filepath.Join(extRoot, localName)
	if err := os.MkdirAll(filepath.Join(localDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	localMF := map[string]any{
		"name":       localName,
		"local":      true,
		"local_path": "/src/bitbottle-greet",
	}
	writeMF(t, filepath.Join(localDir, "manifest.json"), localMF)

	mgr := extensions.New(extRoot, nil)
	exts, err := mgr.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(exts) != 2 {
		t.Fatalf("List returned %d extensions; want 2", len(exts))
	}

	byName := map[string]extensions.Extension{}
	for _, e := range exts {
		byName[e.Name] = e
	}

	if e, ok := byName[remoteName]; !ok {
		t.Errorf("remote extension %q not in list", remoteName)
	} else if e.Version != "v1.2.3" {
		t.Errorf("remote version = %q; want v1.2.3", e.Version)
	}

	if e, ok := byName[localName]; !ok {
		t.Errorf("local extension %q not in list", localName)
	} else if !e.Local {
		t.Errorf("local extension Local = false; want true")
	}
}

// --- Remove tests ---

func TestRemove_HappyPath(t *testing.T) {
	srv := newTestServer(t, "v1.0.0", []byte("binary"))
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := mgr.Remove("hello"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "hello")); !os.IsNotExist(err) {
		t.Errorf("extension dir still exists after Remove")
	}
}

func TestRemove_NotInstalled(t *testing.T) {
	mgr := extensions.New(t.TempDir(), nil)
	if err := mgr.Remove("nonexistent"); err == nil {
		t.Fatal("expected error removing non-existent extension; got nil")
	}
}

// --- Upgrade tests ---

// newUpgradeServer serves two releases: one at /v1 and one at /v2.
// The release endpoint returns whichever tag is set at call time.
func newUpgradeServer(t *testing.T, tag string, binaryContent []byte) (*httptest.Server, func(newTag string)) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	currentTag := tag
	setTag := func(newTag string) { currentTag = newTag }

	mux.HandleFunc("/repos/owner/bitbottle-hello/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(makeRelease(currentTag, srv.URL+"/download"))
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(binaryContent)
	})
	return srv, setTag
}

func TestUpgrade_AlreadyUpToDate(t *testing.T) {
	srv, _ := newUpgradeServer(t, "v1.0.0", []byte("binary"))
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("install: %v", err)
	}

	old, newV, err := mgr.Upgrade("hello", false)
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if old != "v1.0.0" || newV != "v1.0.0" {
		t.Errorf("Upgrade = (%q, %q); want (v1.0.0, v1.0.0)", old, newV)
	}
}

func TestUpgrade_UpgradeAvailable(t *testing.T) {
	srv, setTag := newUpgradeServer(t, "v1.0.0", []byte("binary"))
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Bump the tag so the server now reports v2.0.0.
	setTag("v2.0.0")

	old, newV, err := mgr.Upgrade("hello", false)
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if old != "v1.0.0" || newV != "v2.0.0" {
		t.Errorf("Upgrade = (%q, %q); want (v1.0.0, v2.0.0)", old, newV)
	}
}

func TestUpgrade_LocalSkipped(t *testing.T) {
	// Build a local extension.
	srcBase := t.TempDir()
	extName := "bitbottle-hello"
	srcDir := filepath.Join(srcBase, extName)
	if err := os.MkdirAll(filepath.Join(srcDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	srcBin := filepath.Join(srcDir, "bin", extName)
	if err := os.WriteFile(srcBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	extRoot := t.TempDir()
	mgr := extensions.New(extRoot, nil)
	if err := mgr.InstallLocal(srcDir, false); err != nil {
		t.Fatalf("InstallLocal: %v", err)
	}

	old, newV, err := mgr.Upgrade("hello", false)
	if err != nil {
		t.Fatalf("Upgrade of local: %v", err)
	}
	if old != "" || newV != "" {
		t.Errorf("Upgrade local = (%q, %q); want ('', '')", old, newV)
	}
}

func TestUpgrade_NotInstalled(t *testing.T) {
	mgr := extensions.New(t.TempDir(), nil)
	_, _, err := mgr.Upgrade("nonexistent", false)
	if err == nil {
		t.Fatal("expected error upgrading non-existent extension; got nil")
	}
}

// --- UpgradeAll tests ---

func TestUpgradeAll(t *testing.T) {
	srv, setTag := newUpgradeServer(t, "v1.0.0", []byte("binary"))
	defer srv.Close()

	dir := t.TempDir()
	mgr := newManagerWithDir(dir, srv)

	if err := mgr.InstallFromGitHub("owner/bitbottle-hello", false); err != nil {
		t.Fatalf("install: %v", err)
	}
	setTag("v2.0.0")

	results := mgr.UpgradeAll(false)
	if err := results["hello"]; err != nil {
		t.Errorf("UpgradeAll[hello]: %v", err)
	}

	// Verify manifest was updated.
	mfPath := filepath.Join(dir, "hello", "manifest.json")
	data, err := os.ReadFile(mfPath)
	if err != nil {
		t.Fatalf("manifest not found: %v", err)
	}
	var mf map[string]any
	if err := json.Unmarshal(data, &mf); err != nil {
		t.Fatalf("corrupt manifest: %v", err)
	}
	if mf["version"] != "v2.0.0" {
		t.Errorf("manifest version = %v; want v2.0.0", mf["version"])
	}
}

// --- Exec tests ---

// installFakeExtension writes a minimal manifest and a stub binary so Exec
// believes the extension is installed.
func installFakeExtension(t *testing.T, extRoot, name string) {
	t.Helper()
	binDir := filepath.Join(extRoot, name, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(binDir, name)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	mf := map[string]any{"name": name, "local": true}
	writeMF(t, filepath.Join(extRoot, name, "manifest.json"), mf)
}

func TestExec_NotInstalled(t *testing.T) {
	mgr := extensions.New(t.TempDir(), nil)
	err := mgr.Exec("nonexistent", nil, "tok", "v0")
	if err == nil {
		t.Fatal("expected error for non-installed extension; got nil")
	}
}

func TestExec_RunnerCalledWithCorrectEnv(t *testing.T) {
	extRoot := t.TempDir()
	installFakeExtension(t, extRoot, "hello")

	// Inject a keyring secret into the current process env so the filter is
	// exercised without mutating os.Environ permanently.
	t.Setenv("MY_KEYRING_PASSPHRASE", "secret")
	t.Setenv("BB_TOKEN", "old-token") // should be overwritten

	var gotBin string
	var gotArgs []string
	var gotEnv []string

	stub := func(bin string, args []string, env []string) error {
		gotBin = bin
		gotArgs = args
		gotEnv = env
		return nil
	}

	mgr := extensions.New(extRoot, nil).WithRunner(stub)
	if err := mgr.Exec("hello", []string{"--verbose"}, "new-token", "v1.2.3"); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	// Binary path must end with the extension name.
	if !strings.HasSuffix(gotBin, filepath.Join("hello", "bin", "hello")) {
		t.Errorf("runner bin = %q; want path ending in hello/bin/hello", gotBin)
	}

	// Args forwarded.
	if len(gotArgs) != 1 || gotArgs[0] != "--verbose" {
		t.Errorf("runner args = %v; want [--verbose]", gotArgs)
	}

	// BB_TOKEN must be the injected value.
	var foundToken, foundVersion bool
	for _, kv := range gotEnv {
		if kv == "BB_TOKEN=new-token" {
			foundToken = true
		}
		if kv == "BITBOTTLE_VERSION=v1.2.3" {
			foundVersion = true
		}
		// KEYRING_PASSPHRASE must be stripped.
		if strings.Contains(strings.ToUpper(kv), "KEYRING_PASSPHRASE") {
			t.Errorf("keyring passphrase leaked into env: %q", kv)
		}
	}
	if !foundToken {
		t.Errorf("BB_TOKEN=new-token not found in runner env: %v", gotEnv)
	}
	if !foundVersion {
		t.Errorf("BITBOTTLE_VERSION=v1.2.3 not found in runner env: %v", gotEnv)
	}
}

func TestExec_RunnerExitErrorPropagated(t *testing.T) {
	extRoot := t.TempDir()
	installFakeExtension(t, extRoot, "hello")

	wantErr := &exec.ExitError{}
	stub := func(_ string, _ []string, _ []string) error { return wantErr }

	mgr := extensions.New(extRoot, nil).WithRunner(stub)
	err := mgr.Exec("hello", nil, "", "v0")
	if err != wantErr {
		t.Errorf("Exec error = %v; want the stub ExitError", err)
	}
}

func writeMF(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// rewriteTransport rewrites the host of every outbound request to the test
// server's base URL so that calls to api.github.com land on the test server.
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	base := rt.base
	switch {
	case len(base) > 7 && base[:7] == "http://":
		clone.URL.Scheme = "http"
		clone.URL.Host = base[7:]
	case len(base) > 8 && base[:8] == "https://":
		clone.URL.Scheme = "https"
		clone.URL.Host = base[8:]
	}
	return rt.inner.RoundTrip(clone)
}
