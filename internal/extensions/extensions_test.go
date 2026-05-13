package extensions_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
