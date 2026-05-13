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
	"sync"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/extension"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

// upgradeSrv is an httptest.Server whose release tag can be changed mid-test.
type upgradeSrv struct {
	*httptest.Server
	mu  sync.Mutex
	tag string
}

func newUpgradeSrv(t *testing.T, initialTag string) *upgradeSrv {
	t.Helper()
	us := &upgradeSrv{tag: initialTag}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/bitbottle-hello/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		us.mu.Lock()
		tag := us.tag
		us.mu.Unlock()
		rel := map[string]any{
			"tag_name": tag,
			"assets": []map[string]any{
				{
					"name":                 fmt.Sprintf("bitbottle-hello_%s_%s", runtime.GOOS, runtime.GOARCH),
					"browser_download_url": "", // filled below after server starts
				},
			},
		}
		// We need the server URL for the download link; use a separate handler path.
		rel["assets"].([]map[string]any)[0]["browser_download_url"] = us.URL + "/download"
		data, _ := json.Marshal(rel)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("binary-content"))
	})

	us.Server = httptest.NewServer(mux)
	t.Cleanup(us.Close)
	return us
}

func (us *upgradeSrv) setTag(tag string) {
	us.mu.Lock()
	us.tag = tag
	us.mu.Unlock()
}

// redirectDefaultTransport redirects http.DefaultTransport to us for the test duration.
func redirectDefaultTransport(t *testing.T, us *upgradeSrv) {
	t.Helper()
	orig := http.DefaultTransport
	http.DefaultTransport = &rewriteRT{base: us.URL, inner: us.Client().Transport}
	t.Cleanup(func() { http.DefaultTransport = orig })
}

// preInstallHello installs owner/bitbottle-hello into configDir using the redirected transport.
func preInstallHello(t *testing.T, configDir string) {
	t.Helper()
	f0, out0, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f0.ConfigDir = func() string { return configDir }
	c0 := extension.NewCmdExtension(f0)
	c0.SetOut(out0)
	c0.SetErr(out0)
	c0.SetArgs([]string{"install", "owner/bitbottle-hello"})
	if err := c0.Execute(); err != nil {
		t.Fatalf("pre-install: %v", err)
	}
}

func TestUpgradeCmd_Single(t *testing.T) {
	us := newUpgradeSrv(t, "v1.0.0")
	configDir := t.TempDir()
	extRoot := filepath.Join(configDir, "extensions")

	redirectDefaultTransport(t, us)
	preInstallHello(t, configDir)
	us.setTag("v2.0.0")

	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }
	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"upgrade", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if !strings.Contains(out.String(), "v2.0.0") {
		t.Errorf("output %q missing new version", out.String())
	}

	mfPath := filepath.Join(extRoot, "hello", "manifest.json")
	data, _ := os.ReadFile(mfPath)
	var mf map[string]any
	_ = json.Unmarshal(data, &mf)
	if mf["version"] != "v2.0.0" {
		t.Errorf("manifest version = %v; want v2.0.0", mf["version"])
	}
}

func TestUpgradeCmd_AlreadyUpToDate(t *testing.T) {
	us := newUpgradeSrv(t, "v1.0.0")
	configDir := t.TempDir()

	redirectDefaultTransport(t, us)
	preInstallHello(t, configDir)

	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }
	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"upgrade", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if !strings.Contains(out.String(), "already at") {
		t.Errorf("output %q missing 'already at'", out.String())
	}
}

func TestUpgradeCmd_All(t *testing.T) {
	us := newUpgradeSrv(t, "v1.0.0")
	configDir := t.TempDir()

	redirectDefaultTransport(t, us)
	preInstallHello(t, configDir)
	us.setTag("v3.0.0")

	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }
	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"upgrade", "--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("upgrade --all: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("output %q missing extension name", out.String())
	}
}

func TestUpgradeCmd_NoNameNoAll(t *testing.T) {
	configDir := t.TempDir()
	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"upgrade"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error with no name and no --all; got nil")
	}
}

func TestUpgradeCmd_NonExistent(t *testing.T) {
	configDir := t.TempDir()
	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"upgrade", "no-such-ext"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error upgrading non-existent extension; got nil")
	}
}
