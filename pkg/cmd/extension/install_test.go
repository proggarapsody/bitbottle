package extension_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/extension"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

// releaseJSON builds a GitHub release response whose single asset name
// matches the current GOOS/GOARCH.
func releaseJSON(downloadURL string) []byte {
	rel := map[string]any{
		"tag_name": "v1.0.0",
		"assets": []map[string]any{
			{
				"name":                 fmt.Sprintf("bitbottle-hello_%s_%s", runtime.GOOS, runtime.GOARCH),
				"browser_download_url": downloadURL,
			},
		},
	}
	data, _ := json.Marshal(rel)
	return data
}

// newInstallServer starts a test HTTP server that serves the GitHub release
// API at /repos/owner/bitbottle-hello/releases/latest and the fake binary at
// /download. It also rewrites all outbound requests to itself so that the
// Manager's hard-coded api.github.com calls land here.
func newInstallServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	relJSON := releaseJSON(srv.URL + "/download")
	mux.HandleFunc("/repos/owner/bitbottle-hello/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(relJSON)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("#!/bin/sh\necho hello\n"))
	})
	return srv
}

// rewriteRT rewrites every request's host to the given test server base URL.
type rewriteRT struct {
	base  string
	inner http.RoundTripper
}

func (rt *rewriteRT) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	base := rt.base
	switch {
	case strings.HasPrefix(base, "http://"):
		clone.URL.Scheme = "http"
		clone.URL.Host = base[len("http://"):]
	case strings.HasPrefix(base, "https://"):
		clone.URL.Scheme = "https"
		clone.URL.Host = base[len("https://"):]
	}
	return rt.inner.RoundTrip(clone)
}

func TestInstallCmd_FromGitHub(t *testing.T) {
	srv := newInstallServer(t)
	defer srv.Close()

	configDir := t.TempDir()
	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	// Redirect the Manager's HTTP calls to our test server. The install
	// command creates its own Manager with nil client (http.DefaultClient).
	// We override http.DefaultTransport for the duration of this test.
	origTransport := http.DefaultTransport
	http.DefaultTransport = &rewriteRT{base: srv.URL, inner: origTransport}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"install", "owner/bitbottle-hello"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install from GitHub: %v", err)
	}

	if !strings.Contains(out.String(), "owner/bitbottle-hello") {
		t.Errorf("output %q missing repo name", out.String())
	}

	// Binary must be present.
	binPath := filepath.Join(configDir, "extensions", "hello", "bin", "hello")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("binary not found at %s: %v", binPath, err)
	}
}

func TestInstallCmd_Local(t *testing.T) {
	// Prepare a local source dir.
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

	configDir := t.TempDir()
	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"install", "--local", srcDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install --local: %v", err)
	}
	if !strings.Contains(out.String(), srcDir) {
		t.Errorf("output %q missing source path", out.String())
	}
}

func TestInstallCmd_AlreadyInstalled(t *testing.T) {
	srv := newInstallServer(t)
	defer srv.Close()

	configDir := t.TempDir()
	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	origTransport := http.DefaultTransport
	http.DefaultTransport = &rewriteRT{base: srv.URL, inner: origTransport}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	// First install.
	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"install", "owner/bitbottle-hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install (no --force) must fail.
	cmd2 := extension.NewCmdExtension(f)
	cmd2.SetOut(out)
	cmd2.SetErr(errOut)
	cmd2.SetArgs([]string{"install", "owner/bitbottle-hello"})
	if err := cmd2.Execute(); err == nil {
		t.Fatal("expected error on second install without --force; got nil")
	}
}
